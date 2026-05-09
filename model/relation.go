package model

type HeadwordRelationEdge struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;default:(-);type:bigint"`

	SourceHeadword           string `gorm:"type:text;not null"`
	SourceHeadwordNormalized string `gorm:"type:text;not null;index:idx_headword_relation_edges_source_headword_pos_type,priority:1;uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:1"`
	SourcePOSCode            int    `gorm:"column:source_pos_code;type:integer;not null;check:source_pos_code IN (1,2,3,4);index:idx_headword_relation_edges_source_headword_pos_type,priority:2;uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:2"`
	RelationType             string `gorm:"type:text;not null;check:relation_type IN ('synonym','antonym','hypernym','hyponym','meronym','holonym','similar_to','also_see','derivation','pertainym','domain_topic','domain_region','exemplifies','attribute','entails','causes','event','agent','result','by_means_of','undergoer','instrument','uses','state','property','location','material','vehicle','participle','body_part','destination');index:idx_headword_relation_edges_source_headword_pos_type,priority:3;uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:3"`
	TargetHeadword           string `gorm:"type:text;not null"`
	TargetHeadwordNormalized string `gorm:"type:text;not null;index:idx_headword_relation_edges_target_headword_pos,priority:1;uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:5"`
	TargetPOSCode            int    `gorm:"column:target_pos_code;type:integer;not null;check:target_pos_code IN (1,2,3,4);index:idx_headword_relation_edges_target_headword_pos,priority:2;uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:6"`

	SourceRelationType string `gorm:"type:text;not null;check:source_relation_type IN ('members','antonym','derivation','pertainym','hypernym','mero_part','mero_member','mero_substance','similar','also','domain_topic','domain_region','exemplifies','attribute','entails','causes','event','agent','result','by_means_of','undergoer','instrument','uses','state','property','location','material','vehicle','participle','body_part','destination');uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:4"`
	SourceSynsetID     string `gorm:"type:text;not null;check:source_synset_id <> '';uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:7"`
	TargetSynsetID     string `gorm:"type:text;not null;check:target_synset_id <> '';uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:8"`
	SourceSenseID      string `gorm:"type:text;not null;default:'';uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:9"`
	TargetSenseID      string `gorm:"type:text;not null;default:'';uniqueIndex:idx_headword_relation_edges_unique_evidence,priority:10"`

	ImportRunID int64 `gorm:"type:bigint;autoIncrement:false;not null;index:idx_headword_relation_edges_import_run_id"`

	ImportRun ImportRun `gorm:"foreignKey:ImportRunID;references:ID;constraint:OnDelete:RESTRICT"`
}

func (HeadwordRelationEdge) TableName() string {
	return "headword_relation_edges"
}
