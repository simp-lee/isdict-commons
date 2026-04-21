package model

type PronunciationIPA struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	EntryID      int64  `gorm:"type:bigint;autoIncrement:false;not null;uniqueIndex:idx_pronunciation_ipas_entry_id_accent_code_ipa,priority:1;index:idx_pronunciation_ipas_entry_id_accent_code_display_order,priority:1"`
	AccentCode   string `gorm:"type:text;not null;default:'unknown';uniqueIndex:idx_pronunciation_ipas_entry_id_accent_code_ipa,priority:2;index:idx_pronunciation_ipas_entry_id_accent_code_display_order,priority:2"`
	IPA          string `gorm:"column:ipa;type:text;not null;uniqueIndex:idx_pronunciation_ipas_entry_id_accent_code_ipa,priority:3"`
	IsPrimary    bool   `gorm:"type:boolean;not null;default:false"`
	DisplayOrder int16  `gorm:"type:smallint;not null;default:1;index:idx_pronunciation_ipas_entry_id_accent_code_display_order,priority:3"`

	Entry Entry `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
}

func (PronunciationIPA) TableName() string {
	return "pronunciation_ipas"
}

type PronunciationAudio struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	EntryID       int64  `gorm:"type:bigint;autoIncrement:false;not null;uniqueIndex:idx_pronunciation_audios_entry_id_accent_code_audio_filename,priority:1;index:idx_pronunciation_audios_entry_id_accent_code_display_order,priority:1"`
	AccentCode    string `gorm:"type:text;not null;default:'unknown';uniqueIndex:idx_pronunciation_audios_entry_id_accent_code_audio_filename,priority:2;index:idx_pronunciation_audios_entry_id_accent_code_display_order,priority:2"`
	AudioFilename string `gorm:"column:audio_filename;type:text;not null;uniqueIndex:idx_pronunciation_audios_entry_id_accent_code_audio_filename,priority:3;index:idx_pronunciation_audios_audio_filename"`
	IsPrimary     bool   `gorm:"type:boolean;not null;default:false"`
	DisplayOrder  int16  `gorm:"type:smallint;not null;default:1;index:idx_pronunciation_audios_entry_id_accent_code_display_order,priority:3"`

	Entry Entry `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
}

func (PronunciationAudio) TableName() string {
	return "pronunciation_audios"
}
