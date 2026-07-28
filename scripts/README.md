# PostgreSQL 备份与恢复

脚本只使用 PostgreSQL 自带的 `pg_dump` 和 `pg_restore`，默认读取标准连接环境变量：
`PGHOST`、`PGPORT`、`PGUSER`、`PGPASSWORD`、`PGDATABASE`。默认数据库名为 `imaiplay`。

## 手动备份

```bash
chmod +x scripts/backup.sh scripts/restore.sh
BACKUP_DIR=./backups RETENTION_COUNT=7 PGDATABASE=imaiplay ./scripts/backup.sh
```

备份采用 custom format，文件名包含 UTC 时间戳。`RETENTION_COUNT` 保留最新 N 份，默认 7 份；
设为 0 或非数字会直接失败。部署前可用 `DRY_RUN=1` 检查命令而不连接数据库：

```bash
DRY_RUN=1 BACKUP_DIR=./backups RETENTION_COUNT=7 ./scripts/backup.sh
```

## 恢复

恢复会使用 `--clean --if-exists` 覆盖目标库对象，必须显式输入 `RESTORE`：

```bash
PGDATABASE=imaiplay ./scripts/restore.sh ./backups/imaiplay_20260728T000000Z.dump
```

先用 `DRY_RUN=1` 查看命令；恢复前应停止应用写入并确认目标数据库。恢复失败时脚本返回非零退出码。

## cron 与保留策略

部署用户可配置每日凌晨 3 点执行，日志交给 cron：

```cron
0 3 * * * cd /srv/imaiplay && BACKUP_DIR=/srv/imaiplay/backups RETENTION_COUNT=14 PGDATABASE=imaiplay ./scripts/backup.sh >> /var/log/imaiplay-backup.log 2>&1
```

建议至少保留 7–14 份日备份，并根据数据库增长和恢复目标定期调整。脚本不负责异机、异地或对象存储归档。

## 恢复演练

1. 使用 `DRY_RUN=1` 检查恢复命令。
2. 从备份服务器复制一份 dump 到隔离环境。
3. 创建临时 PostgreSQL 数据库，执行 restore 并输入 `RESTORE`。
4. 启动应用指向临时库，检查 `/health/db`、登录、课程和资源元数据。
5. 记录恢复耗时、数据时间点和缺失项，演练结束后删除临时库。
