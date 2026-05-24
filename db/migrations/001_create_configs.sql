-- 创建配置表
CREATE TABLE IF NOT EXISTS `configs` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `namespace`   VARCHAR(64)     NOT NULL DEFAULT 'default' COMMENT '命名空间，用于环境隔离',
  `group`       VARCHAR(128)    NOT NULL DEFAULT 'DEFAULT_GROUP' COMMENT '配置分组',
  `data_id`     VARCHAR(256)    NOT NULL COMMENT '配置项唯一标识',
  `content`     MEDIUMTEXT      NOT NULL COMMENT '配置内容（JSON 或纯文本）',
  `content_md5` CHAR(32)        NOT NULL COMMENT '内容 MD5，用于变更检测',
  `format`      VARCHAR(16)     NOT NULL DEFAULT 'json' COMMENT '格式：json / text',
  `description` VARCHAR(512)    DEFAULT '' COMMENT '配置描述',
  `version`     BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '版本号，每次修改递增',
  `is_deleted`  TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '软删除标记',
  `created_by`  VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '创建人',
  `updated_by`  VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '最后修改人',
  `created_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_namespace_group_dataid` (`namespace`, `group`, `data_id`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配置主表';