# Copyright (c) HashiCorp, Inc.

# Secrets can be imported - or rather, the key can be overwritten with specified value - by providing store_id.key as the resource id.
terraform import storytelfastly_secretstore_entry.entry abcdef0123456789.my_secret_name


