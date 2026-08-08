package service

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/1622359590/imaiplay/internal/domain"
	"github.com/1622359590/imaiplay/internal/repository"
)

const (
	DomainStateNone                = "none"
	DomainStatePendingVerification = "pending_verification"
	DomainStateVerified            = "verified"
	DomainStateCreatingSite        = "creating_site"
	DomainStateConfiguring         = "configuring"
	DomainStateReady               = "ready"
	DomainStateVerificationFailed  = "verification_failed"
	DomainStateSetupFailed         = "setup_failed"
)

const disableAdminSnippet = `# ImaiPlay tenant admin guard
location ^~ /admin {
    return 404;
}`

type DomainPanel interface {
	AddSite(domain string) (int, error)
	AddReverseProxy(domain, proxyName, target string) error
	ApplyLetsEncrypt(domain string) error
	EnableHTTPSRedirect(domain string) error
	AddNginxSnippet(domain, snippet string) error
	DeleteSite(domain string) error
	GetSiteInfo(domain string) (map[string]interface{}, error)
}

type DomainResolver interface {
	LookupCNAME(host string) (string, error)
	LookupIP(host string) ([]net.IP, error)
}

type DomainAuditRecorder interface {
	Record(context.Context, domain.AuditEvent) error
}

type DomainBindConfig struct {
	ExpectedIP     string
	ReservedDomain string
	CNAMETarget    string
	ProxyTarget    string
	PollInterval   time.Duration
	MaxPolls       int
}

type DomainBindStatus struct {
	State            string    `json:"state"`
	Domain           string    `json:"domain,omitempty"`
	Message          string    `json:"message,omitempty"`
	CurrentStep      int       `json:"current_step"`
	TotalSteps       int       `json:"total_steps"`
	CNAMETarget      string    `json:"cname_target"`
	TenantCode       string    `json:"tenant_code,omitempty"`
	DefaultPortalURL string    `json:"default_portal_url,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
	History          []string  `json:"-"`
}

type DomainBindService struct {
	tenants  repository.TenantRepository
	panel    DomainPanel
	resolver DomainResolver
	audit    DomainAuditRecorder
	config   DomainBindConfig

	mu       sync.RWMutex
	statuses map[string]DomainBindStatus
	owners   map[string]string
}

type netDomainResolver struct{}

func (netDomainResolver) LookupCNAME(host string) (string, error) {
	return net.LookupCNAME(host)
}

func (netDomainResolver) LookupIP(host string) ([]net.IP, error) {
	return net.LookupIP(host)
}

func NewDomainBindService(
	tenants repository.TenantRepository,
	panel DomainPanel,
	resolver DomainResolver,
	audit DomainAuditRecorder,
	config DomainBindConfig,
) *DomainBindService {
	if resolver == nil {
		resolver = netDomainResolver{}
	}
	if config.ReservedDomain == "" {
		config.ReservedDomain = "play.imai.work"
	}
	if config.CNAMETarget == "" {
		config.CNAMETarget = config.ReservedDomain
	}
	if config.ProxyTarget == "" {
		config.ProxyTarget = "http://127.0.0.1:18080"
	}
	if config.PollInterval <= 0 {
		config.PollInterval = 5 * time.Second
	}
	if config.MaxPolls <= 0 {
		config.MaxPolls = 24
	}
	return &DomainBindService{
		tenants: tenants, panel: panel, resolver: resolver, audit: audit,
		config: config, statuses: make(map[string]DomainBindStatus),
		owners: make(map[string]string),
	}
}
