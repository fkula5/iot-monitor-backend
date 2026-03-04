package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// AlertRule holds the schema definition for the AlertRule entity.
type AlertRule struct {
	ent.Schema
}

// Fields of the AlertRule.
func (AlertRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.Int64("sensor_id"),
		field.String("condition_type").
			Default("GT"),
		field.Float("threshold"),
		field.String("description").Optional(),
		field.Bool("is_enabled").Default(true),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the AlertRule.
func (AlertRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("alerts", Alert.Type),
	}
}
