package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SensorGroup holds the schema definition for the SensorGroup entity.
type SensorGroup struct {
	ent.Schema
}

// Fields of the SensorGroup.
func (SensorGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("Name of the sensor group"),
		field.String("description").
			Optional().
			Comment("Description of the sensor group"),
		field.String("color").
			Optional().
			Default("#3B82F6").
			Comment("Color for UI display (hex format)"),
		field.String("icon").
			Optional().
			Comment("Icon identifier for UI"),
		field.Int64("user_id").
			Comment("ID of the user who owns the group"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the SensorGroup.
func (SensorGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sensors", Sensor.Type).
			Comment("Sensors in this group"),
	}
}
