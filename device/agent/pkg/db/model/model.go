package model

var models = make([]interface{}, 0)

func register(model ...interface{}) {
	models = append(models, model...)
}

// GetMigrationModels is a helper function returns all models for table initalization.
func GetMigrationModels() []interface{} {
	return models
}
