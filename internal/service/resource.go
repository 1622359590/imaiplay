package service

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	usercontext "github.com/1622359590/imaiplay/internal/context"
	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/errorsx"
	"github.com/1622359590/imaiplay/internal/repository"
	"github.com/1622359590/imaiplay/internal/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxResourceSize int64 = 500 * 1024 * 1024

var resourceMIME = map[string]struct {
	resourceType string
	extension    string
}{
	"image/jpeg":      {"image", ".jpg"},
	"image/png":       {"image", ".png"},
	"image/webp":      {"image", ".webp"},
	"video/mp4":       {"video", ".mp4"},
	"video/webm":      {"video", ".webm"},
	"application/pdf": {"document", ".pdf"},
}

type ResourceService struct {
	resources repository.ResourceRepository
	storage   storage.Storage
	quota     interface {
		CheckStorage(context.Context, string, int64) error
	}
}

func NewResourceService(
	resources repository.ResourceRepository, fileStorage storage.Storage, quota ...interface {
		CheckStorage(context.Context, string, int64) error
	},
) *ResourceService {
	service := &ResourceService{resources: resources, storage: fileStorage}
	if len(quota) > 0 {
		service.quota = quota[0]
	}
	return service
}

func (service *ResourceService) Upload(
	ctx context.Context, name string, reader io.Reader, size int64,
) (*domain.Resource, error) {
	userID, tenantID, err := resourceManager(ctx)
	if err != nil {
		return nil, err
	}
	if size <= 0 || size > maxResourceSize {
		return nil, unsupportedResource()
	}
	if service.quota != nil {
		if err := service.quota.CheckStorage(ctx, tenantID, size); err != nil {
			return nil, err
		}
	}
	buffered := bufio.NewReader(reader)
	prefix, err := buffered.Peek(minInt64(size, 512))
	if err != nil && err != bufio.ErrBufferFull && err != io.EOF {
		return nil, unsupportedResource()
	}
	format, ok := detectResourceFormat(prefix)
	if !ok {
		return nil, unsupportedResource()
	}
	key := tenantID + "/" + uuid.NewString() + format.extension
	url, err := service.storage.Put(ctx, key, buffered, size)
	if err != nil {
		return nil, errorsx.Internal("store resource failed")
	}
	resource := &domain.Resource{
		BaseModel: domain.BaseModel{TenantID: tenantID},
		Name:      strings.TrimSpace(name), ResourceType: format.resourceType,
		URL: url, SizeBytes: size, CreatedBy: userID,
	}
	if resource.Name == "" {
		resource.Name = uuid.NewString() + format.extension
	}
	if err := service.resources.Create(ctx, resource); err != nil {
		_ = service.storage.Delete(ctx, key)
		return nil, errorsx.Internal("create resource failed")
	}
	return resource, nil
}

func (service *ResourceService) List(
	ctx context.Context, offset, limit int,
) ([]domain.Resource, int64, error) {
	_, tenantID, err := resourceManager(ctx)
	if err != nil {
		return nil, 0, err
	}
	items, total, err := service.resources.FindByTenant(
		ctx, tenantID, offset, limit,
	)
	if err != nil {
		return nil, 0, errorsx.Internal("list resources failed")
	}
	return items, total, nil
}

func (service *ResourceService) File(
	ctx context.Context, id, storageRoot string,
) (path, contentType, fileName string, err error) {
	_, tenantID, _, _, ok := usercontext.UserFromContext(ctx)
	if !ok || tenantID == "" {
		return "", "", "", errorsx.Unauthorized("missing or invalid token")
	}
	resource, err := service.resources.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) || err != nil {
		return "", "", "", errorsx.NotFound("resource not found")
	}
	if resource.TenantID != tenantID {
		return "", "", "", errorsx.NotFound("resource not found")
	}
	key, err := service.storageKey(resource.URL)
	if err != nil {
		return "", "", "", errorsx.NotFound("resource not found")
	}
	path, err = safeResourcePath(storageRoot, key)
	if err != nil {
		return "", "", "", errorsx.NotFound("resource not found")
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", "", "", errorsx.NotFound("resource not found")
	}
	contentType = mime.TypeByExtension(filepath.Ext(key))
	if contentType == "" {
		contentType = resourceContentType(resource.ResourceType)
	}
	return path, contentType, resource.Name, nil
}

// Open applies the same tenant authorization as File and streams either local
// or object-storage content without exposing a public object URL.
func (service *ResourceService) Open(ctx context.Context, id string) (io.ReadCloser, string, string, error) {
	_, tenantID, _, _, ok := usercontext.UserFromContext(ctx)
	if !ok || tenantID == "" {
		return nil, "", "", errorsx.Unauthorized("missing or invalid token")
	}
	resource, err := service.resources.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) || err != nil || resource.TenantID != tenantID {
		return nil, "", "", errorsx.NotFound("resource not found")
	}
	key, err := service.storageKey(resource.URL)
	if err != nil {
		return nil, "", "", errorsx.NotFound("resource not found")
	}
	readable, ok := service.storage.(interface {
		Get(context.Context, string) (io.ReadCloser, error)
	})
	if !ok {
		return nil, "", "", errorsx.NotFound("resource not found")
	}
	body, err := readable.Get(ctx, key)
	if err != nil {
		return nil, "", "", errorsx.NotFound("resource not found")
	}
	contentType := mime.TypeByExtension(filepath.Ext(key))
	if contentType == "" {
		contentType = resourceContentType(resource.ResourceType)
	}
	return body, contentType, resource.Name, nil
}

func (service *ResourceService) Delete(ctx context.Context, id string) error {
	if _, _, err := resourceManager(ctx); err != nil {
		return err
	}
	resource, err := service.resources.FindByID(ctx, id)
	if err != nil {
		return mapNotFound(err, "resource not found")
	}
	key, err := service.storageKey(resource.URL)
	if err != nil {
		return err
	}
	if err := service.resources.Delete(ctx, id); err != nil {
		return mapNotFound(err, "resource not found")
	}
	if err := service.storage.Delete(ctx, key); err != nil {
		if restoreErr := service.resources.Create(ctx, resource); restoreErr != nil {
			return errorsx.Internal("delete resource failed and restore record failed")
		}
		return errorsx.Internal("delete resource file failed")
	}
	return nil
}

func (service *ResourceService) storageKey(url string) (string, error) {
	base := strings.TrimRight(service.storage.URL(""), "/")
	prefix := base + "/"
	if base == "" || !strings.HasPrefix(url, prefix) {
		return "", errorsx.Internal("invalid resource URL")
	}
	key := strings.TrimPrefix(url, prefix)
	if key == "" {
		return "", errorsx.Internal("invalid resource URL")
	}
	return key, nil
}

func safeResourcePath(storageRoot, key string) (string, error) {
	if strings.TrimSpace(storageRoot) == "" || key == "" || strings.Contains(key, "..") {
		return "", errors.New("invalid resource path")
	}
	root, err := filepath.Abs(storageRoot)
	if err != nil {
		return "", err
	}
	cleanKey := filepath.Clean(filepath.FromSlash(key))
	if cleanKey == "." || filepath.IsAbs(cleanKey) {
		return "", errors.New("invalid resource path")
	}
	path := filepath.Clean(filepath.Join(root, cleanKey))
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("invalid resource path")
	}
	return path, nil
}

func resourceContentType(resourceType string) string {
	switch resourceType {
	case "image":
		return "image/*"
	case "video":
		return "video/*"
	case "document":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

func resourceManager(ctx context.Context) (string, string, error) {
	userID, tenantID, _, role, ok := usercontext.UserFromContext(ctx)
	if !ok || role != "tenant_admin" || userID == "" || tenantID == "" {
		return "", "", errorsx.Forbidden("permission denied")
	}
	return userID, tenantID, nil
}

func unsupportedResource() error {
	return errorsx.BadRequest("unsupported file type or size exceeds limit")
}

func detectResourceFormat(prefix []byte) (struct {
	resourceType string
	extension    string
}, bool) {
	detected := http.DetectContentType(prefix)
	if format, ok := resourceMIME[detected]; ok {
		return format, true
	}
	if len(prefix) >= 12 && bytes.Equal(prefix[4:8], []byte("ftyp")) {
		return resourceMIME["video/mp4"], true
	}
	if len(prefix) >= 4 && bytes.Equal(
		prefix[:4], []byte{0x1a, 0x45, 0xdf, 0xa3},
	) {
		return resourceMIME["video/webm"], true
	}
	return struct {
		resourceType string
		extension    string
	}{}, false
}

func minInt64(left, right int64) int {
	if left < right {
		return int(left)
	}
	return int(right)
}
