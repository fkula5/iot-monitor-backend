package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// SensorType holds the schema definition for the SensorType entity.
type SensorType struct {
	ent.Schema
}

// Fields of the SensorType.
func (SensorType) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique().
			Comment("Name of the sensor type"),
		field.String("model").
			NotEmpty().
			Comment("Model of the sensor type"),
		field.String("manufacturer").
			Optional().
			Comment("Manufacturer of the sensor type"),
		field.String("description").
			Optional().
			Comment("Description of the sensor type"),
		field.String("unit").Optional(),
		field.Float("min_value").Optional(),
		field.Float("max_value").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the SensorType.
func (SensorType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sensors", Sensor.Type),
	}
}
