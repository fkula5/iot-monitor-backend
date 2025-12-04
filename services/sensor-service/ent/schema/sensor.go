package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Sensor holds the schema definition for the Sensor entity.
type Sensor struct {
	ent.Schema
}

// Fields of the Sensor.
func (Sensor) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("Name of the sensor"),
		field.String("location").
			Optional().
			Comment("Physical location of the sensor"),
		field.String("description").
			Optional().
			Comment("Description of the sensor"),
		field.Bool("active").
			Default(true).
			Comment("Whether the sensor is currently active"),
		field.Time("last_updated").
			Optional().
			Comment("Last time the sensor value was updated"),
		field.Int64("user_id").
			Comment("ID of the user who owns the sensor"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Time when the sensor was created"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Sensor.
func (Sensor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("type", SensorType.Type).
			Unique().
			Required().
			Comment("The type of this sensor"),
	}
}
