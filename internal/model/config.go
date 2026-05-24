package model

import "time"

type Config struct {
	ID          uint64    `db:"id" json:"id"`
	Namespace   string    `db:"namespace" json:"namespace"`
	Group       string    `db:"group" json:"group"`
	DataID      string    `db:"data_id" json:"dataId"`
	Content     string    `db:"content" json:"content"`
	ContentMD5  string    `db:"content_md5" json:"contentMd5"`
	Format      string    `db:"format" json:"format"`
	Description string    `db:"description" json:"description"`
	Version     uint64    `db:"version" json:"version"`
	IsDeleted   bool      `db:"is_deleted" json:"-"`
	CreatedBy   string    `db:"created_by" json:"createdBy"`
	UpdatedBy   string    `db:"updated_by" json:"updatedBy"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

type ConfigHistory struct {
	ID         uint64    `db:"id" json:"id"`
	ConfigID   uint64    `db:"config_id" json:"configId"`
	Namespace  string    `db:"namespace" json:"namespace"`
	Group      string    `db:"group" json:"group"`
	DataID     string    `db:"data_id" json:"dataId"`
	Content    string    `db:"content" json:"content"`
	ContentMD5 string    `db:"content_md5" json:"contentMd5"`
	Version    uint64    `db:"version" json:"version"`
	OpType     string    `db:"op_type" json:"opType"`
	OpBy       string    `db:"op_by" json:"opBy"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
}

type ServerNode struct {
	ID        uint64    `db:"id" json:"id"`
	NodeID    string    `db:"node_id" json:"nodeId"`
	Address   string    `db:"address" json:"address"`
	Status    string    `db:"status" json:"status"`
	LastBeat  time.Time `db:"last_beat" json:"lastBeat"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}