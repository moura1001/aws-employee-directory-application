package utils

import "os"

var PHOTOS_BUCKET = os.Getenv("PHOTOS_BUCKET")
var CSRF_SECRET = "32-byte-long-auth-key"

var DATABASE_HOST = os.Getenv("DATABASE_HOST")
var DATABASE_USER = os.Getenv("DATABASE_USER")
var DATABASE_PASSWORD = os.Getenv("DATABASE_PASSWORD")
var DATABASE_DB_NAME = os.Getenv("DATABASE_DB_NAME")

var DYNAMO_MODE = os.Getenv("DYNAMO_MODE")
