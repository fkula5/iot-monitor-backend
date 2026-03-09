package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Alert struct {
	ent.Schema
}

func (Alert) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		field.Float("value"),
		field.String("message"),
		field.Time("triggered_at").Default(time.Now),
		field.Bool("is_read").Default(false),
	}
}

func (Alert) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("rule", AlertRule.Type).
			Ref("alerts").
			Unique().
			Required(),
	}
}
