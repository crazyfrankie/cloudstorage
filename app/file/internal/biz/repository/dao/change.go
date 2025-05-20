package dao

// ChangeOperation 表示变更操作的类型
type ChangeOperation string

const (
	Insert ChangeOperation = "insert" // 插入内容
	Delete ChangeOperation = "delete" // 删除内容
	Update ChangeOperation = "update" // 更新内容
)

// FileChange 表示文件的变更记录
type FileChange struct {
	ID        int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	FileID    int64           `json:"file_id" gorm:"column:file_id;index"`
	Version   int64           `json:"version" gorm:"column:version"`
	Operation ChangeOperation `json:"operation" gorm:"column:operation"`
	Position  int64           `json:"position" gorm:"column:position"` // 变更的起始位置（字节偏移量）
	Length    int64           `json:"length" gorm:"column:length"`     // 原内容长度（用于Delete和Update操作）
	Content   []byte          `json:"content" gorm:"column:content"`   // 新内容（用于Insert和Update操作）
	DeviceID  string          `json:"device_id" gorm:"column:device_id"`
	CreatedAt int64           `json:"created_at" gorm:"column:created_at"`
}

// TableName 指定数据库表名
func (FileChange) TableName() string {
	return "file_changes"
}
