package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("username").
			Unique().
			NotEmpty(),
		field.String("password_hash").
			Sensitive().NotEmpty(),
		field.String("first_name").
			Optional(),
		field.String("last_name").
			Optional(),
		field.Bool("active").
			Default(true),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.String("refresh_token").
			Optional().
			Sensitive(),
		field.Time("refresh_token_expires").
			Optional(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
