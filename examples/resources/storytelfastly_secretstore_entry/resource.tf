# Copyright (c) HashiCorp, Inc.

resource "storytelfastly_secretstore_entry" "entry" {
  store_id = "myid"
  key      = "mykey"
  value    = "my-secret-value"
}
