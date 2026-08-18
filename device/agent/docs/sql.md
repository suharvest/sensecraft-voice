# 创建 `sensecraftVoice` 数据库

```sql
CREATE DATABASE sensecraftVoice;
```

## 创建 `audit`

```sql
CREATE TABLE `audits`
(
    `id`               int primary key                  NOT NULL AUTO_INCREMENT COMMENT '主键',
    `gmt_create`       datetime COMMENT '创建时间',
    `gmt_modified`     datetime COMMENT '修改时间',
    `resource_version` int COMMENT '版本号',
    `operator`         varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '操作人',
    `action`           varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '动作',
    `ip`               varchar(128) COLLATE utf8mb4_bin NOT NULL COMMENT '来源ip',
    `status`           tinyint(4) COLLATE utf8mb4_bin NOT NULL COMMENT '执行是否成功：0-失败，1-成功',
    `path`             varchar(255) COLLATE utf8mb4_bin NOT NULL COMMENT '详细内容',
    `resource_type`    varchar(128) COLLATE utf8mb4_bin NOT NULL COMMENT '操作的资源类型'
) ENGINE=InnoDB AUTO_INCREMENT=3355 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin
```