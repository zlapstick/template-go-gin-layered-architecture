package seeders

import "database/sql"

func init() {
	Register("000002_third_party_keys_seeder", Seed000002ThirdPartyKeysSeeder, Rollback000002ThirdPartyKeysSeeder)
}

func Seed000002ThirdPartyKeysSeeder(db *sql.DB) error {
	query := "INSERT INTO third_party_keys (name, key) VALUES ($1, $2)"

	keys := []struct {
		Name  string
		Key   string
	}{
		{"smartchat", "aW35JuFJyjFQCNdHhTQ/JAozrCqGRDeqnwp+E4hZ1bM="},
	}

	for _, k := range keys {
		if _, err := db.Exec(query, k.Name, k.Key); err != nil {
			return err
		}
	}
	return nil
}

func Rollback000002ThirdPartyKeysSeeder(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM third_party_keys WHERE name IN ($1)", "smartchat")
	return err
}
