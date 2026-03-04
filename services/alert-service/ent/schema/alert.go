package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Alert holds the schema definition for the Alert entity.
type Alert struct {
	ent.Schema
}

// Fields of the Alert.
func (Alert) Fields() []ent.Field {
	return []ent.Field{
		field.Float("value"),
		field.String("message"),
		field.Time("triggered_at").Default(time.Now),
		field.Bool("is_read").Default(false),
	}
}

// Edges of the Alert.
func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("rule", AlertRule.Type).
			Ref("alerts").
			Unique().
			Required(),
	}
}
