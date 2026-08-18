package db

import (
	"gorm.io/gorm"
)

type ShareDaoFactory interface {
	Audit() AuditInterface
	Device() DeviceInterface
	Recording() RecordingInterface
	Store() StoreInterface
	Location() LocationInterface
	User() UserInterface
	Stats() StatsInterface
	Chat() ChatInterface
	FileUpload() FileUploadInterface
	AudioRecording() AudioRecordingInterface
	SystemPrompt() SystemPromptInterface
	Keyword() KeywordInterface
	KeywordMatch() KeywordMatchInterface
	QueryLog() QueryLogInterface
	SeeedAPILog() SeeedAPILogInterface
	AsrServer() AsrServerInterface
}

type shareDaoFactory struct {
	db *gorm.DB
}

func (f *shareDaoFactory) Audit() AuditInterface                   { return newAudit(f.db) }
func (f *shareDaoFactory) Device() DeviceInterface                 { return newDevice(f.db) }
func (f *shareDaoFactory) Recording() RecordingInterface           { return newRecording(f.db) }
func (f *shareDaoFactory) Store() StoreInterface                   { return newStore(f.db) }
func (f *shareDaoFactory) Location() LocationInterface             { return newLocation(f.db) }
func (f *shareDaoFactory) User() UserInterface                     { return newUser(f.db) }
func (f *shareDaoFactory) Stats() StatsInterface                   { return newStats(f.db) }
func (f *shareDaoFactory) Chat() ChatInterface                     { return NewChat(f.db) }
func (f *shareDaoFactory) FileUpload() FileUploadInterface         { return newFileUpload(f.db) }
func (f *shareDaoFactory) AudioRecording() AudioRecordingInterface { return newAudioRecording(f.db) }
func (f *shareDaoFactory) SystemPrompt() SystemPromptInterface     { return newSystemPrompt(f.db) }
func (f *shareDaoFactory) Keyword() KeywordInterface               { return newKeyword(f.db) }
func (f *shareDaoFactory) KeywordMatch() KeywordMatchInterface     { return newKeywordMatch(f.db) }
func (f *shareDaoFactory) QueryLog() QueryLogInterface             { return newQueryLog(f.db) }
func (f *shareDaoFactory) SeeedAPILog() SeeedAPILogInterface       { return newSeeedAPILog(f.db) }
func (f *shareDaoFactory) AsrServer() AsrServerInterface           { return newAsrServer(f.db) }

func NewDaoFactory(db *gorm.DB, migrate bool) (ShareDaoFactory, error) {
	if migrate {
		// 自动创建指定模型的数据库表结构
		if err := newMigrator(db).AutoMigrate(); err != nil {
			return nil, err
		}
	}

	return &shareDaoFactory{
		db: db,
	}, nil
}
