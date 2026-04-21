package model

import "github.com/lib/pq"

type EntryForm struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	EntryID         int64          `gorm:"type:bigint;autoIncrement:false;not null;index:idx_entry_forms_entry_id_relation_kind,priority:1"`
	FormText        string         `gorm:"type:text;not null"`
	NormalizedForm  string         `gorm:"type:text;not null;index:idx_entry_forms_normalized_form"`
	RelationKind    string         `gorm:"type:text;not null;check:relation_kind IN ('form','alias');index:idx_entry_forms_entry_id_relation_kind,priority:2"`
	FormType        *string        `gorm:"type:text"`
	SourceRelations pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	DisplayOrder    int16          `gorm:"type:smallint;not null;default:1"`

	Entry Entry `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
}

func (EntryForm) TableName() string {
	return "entry_forms"
}

type LexicalRelation struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	EntryID              int64  `gorm:"type:bigint;autoIncrement:false;not null;index:idx_lexical_relations_entry_id_relation_type,priority:1"`
	SenseID              *int64 `gorm:"type:bigint;autoIncrement:false;index:idx_lexical_relations_sense_id_relation_type,priority:1"`
	RelationType         string `gorm:"type:text;not null;check:relation_type IN ('synonym','antonym','derived');index:idx_lexical_relations_entry_id_relation_type,priority:2;index:idx_lexical_relations_sense_id_relation_type,priority:2"`
	TargetText           string `gorm:"type:text;not null"`
	TargetTextNormalized string `gorm:"type:text;not null"`
	DisplayOrder         int16  `gorm:"type:smallint;not null;default:1"`

	Entry Entry  `gorm:"foreignKey:EntryID;references:ID;constraint:OnDelete:CASCADE"`
	Sense *Sense `gorm:"foreignKey:SenseID;references:ID;constraint:OnDelete:CASCADE"`
}

func (LexicalRelation) TableName() string {
	return "lexical_relations"
}
