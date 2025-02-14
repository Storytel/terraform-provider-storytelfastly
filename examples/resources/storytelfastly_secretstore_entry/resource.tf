# Copyright (c) Storytel AB

resource "storytelfastly_secretstore_entry" "entry" {
  store_id = "myid"
  key      = "mykey"
  value    = "my-secret-value"
}
