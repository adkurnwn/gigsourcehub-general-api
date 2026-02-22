schema "public" {
}

table "users" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "name" {
    type = varchar(255)
    null = false
  }
  column "email" {
    type = varchar(255)
    null = false
  }
  column "password" {
    type = varchar(255)
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
  }
  column "updated_at" {
    type = timestamptz
    null = false
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }

  column "pendidikan_terakhir" { type = varchar(255); null = true }
  column "instansi_pendidikan" { type = varchar(255); null = true }
  column "jurusan" { type = varchar(255); null = true }
  column "ipk" { type = varchar(50); null = true }
  column "kabupaten" { type = varchar(255); null = true }
  column "provinsi" { type = varchar(255); null = true }
  column "lama_pengalaman_kerja" { type = varchar(255); null = true }
  column "bidang_minat" { type = varchar(255); null = true }
  column "applied_role" { type = varchar(255); null = true }
  column "skills" { type = jsonb; null = true }
  column "link_portofolio" { type = varchar(255); null = true }

  primary_key {
    columns = [column.id]
  }

  index "idx_users_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "cvs" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "name" {
    type = varchar(255)
    null = false
  }
  column "path" {
    type = varchar(255)
    null = false
  }
  column "parsed_data" {
    type = jsonb
    null = true
  }
  column "status" {
    type = varchar(50)
    default = "'UPLOADED'"
    null = true
  }
  column "created_at" {
    type = timestamptz
    null = false
  }
  column "updated_at" {
    type = timestamptz
    null = false
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }

  column "user_id" {
    type = uuid
    null = false
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_cvs_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_cvs_user_id" {
    columns = [column.user_id]
    unique  = true
  }
}

table "sectors" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "name" {
    type = varchar(255)
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
  }
  column "updated_at" {
    type = timestamptz
    null = false
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_sectors_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "roles" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "name" {
    type = varchar(255)
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
  }
  column "updated_at" {
    type = timestamptz
    null = false
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }

  column "sector_id" {
    type = uuid
    null = false
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_roles_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_roles_sector_id" {
    columns = [column.sector_id]
  }
}