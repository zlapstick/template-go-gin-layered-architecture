package config

import (
	"log"
	"os"
)

type BunnyConfig struct {
	BunnyStorageAccessKey string
	BunnyStorageZoneName  string
	BunnyStorageLocation  string
	BunnyStorageCDNUrl    string
}

func BunnyLoad() *BunnyConfig {
	apiKey := os.Getenv("BUNNY_STORAGE_ACCESS_KEY")
	if apiKey == "" {
		log.Fatal("gagal memuat konfigurasi: BUNNY_STORAGE_ACCESS_KEY is not set")
	}

	storageZoneName := os.Getenv("BUNNY_STORAGE_ZONE_NAME")
	if storageZoneName == "" {
		storageZoneName = "cms-byb"
	}

	storageLocation := os.Getenv("BUNNY_STORAGE_LOCATION")
	if storageLocation == "" {
		log.Fatal("gagal memuat konfigurasi: BUNNY_STORAGE_LOCATION is not set")
	}

	cdnUrl := os.Getenv("BUNNY_STORAGE_CDN_BASE_URL")
	if cdnUrl == "" {
		log.Fatal("gagal memuat konfigurasi: BUNNY_STORAGE_CDN_BASE_URL is not set")
	}

	return &BunnyConfig{
		BunnyStorageAccessKey: apiKey,
		BunnyStorageZoneName:  storageZoneName,
		BunnyStorageLocation:  storageLocation,
		BunnyStorageCDNUrl:    cdnUrl,
	}
}
