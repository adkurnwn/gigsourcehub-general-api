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
  column "birthdate" {
    type = date
    null = true
  }
  column "school_university" {
    type = varchar(255)
    null = true
  }
  column "major" {
    type = varchar(255)
    null = true
  }
  column "gpa" {
    type = decimal(3,2)
    null = true
  }
  column "cv_id" {
    type = uuid
    null = true
  }
  column "phone_number" {
    type = varchar(20)
    null = true
  }
  column "portofolio_link" {
    type = text
    null = true
  }
  column "kabupaten_kota_id" {
    type = varchar(5)
    null = true
  }
  column "years_experience" {
    type = int
    null = true
  }
  column "tech_stack" {
    type = jsonb
    null = true
  }
  column "profile_picture" {
    type = varchar(255)
    null = true
  }
  column "candidate_level" {
    type = varchar(50)
    null = true
  }
  column "recruitment_status_id" {
    type = uuid
    null = true
  }
  column "unavailable_until" {
    type = date
    null = true
  }
  column "role_system_id" {
    type = uuid
    null = true
  }
  column "role_applied_id" {
    type = uuid
    null = true
  }
  column "account_status" {
    type = varchar(10)
    null = true
  }
  column "jabatan" {
    type = varchar(100)
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

  primary_key {
    columns = [column.id]
  }

  index "idx_users_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_users_kabupaten_kota_id" {
    columns = [column.kabupaten_kota_id]
  }
  
  index "idx_users_role_applied_id" {
    columns = [column.role_applied_id]
  }
}

table "cvs" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "filename" {
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
    type = varchar(150)
    null = false
  }
  column "is_active" {
    type    = boolean
    default = true
    null    = false
  }
  column "hex_code" {
    type = varchar(10)
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

  primary_key {
    columns = [column.id]
  }

  index "idx_sectors_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "role_system" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "name" {
    type = varchar(150)
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

  index "idx_role_system_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "role_applied" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "name" {
    type = varchar(150)
    null = false
  }
  column "sector_id" {
    type = uuid
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

  index "idx_role_applied_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_role_applied_sector_id" {
    columns = [column.sector_id]
  }
}

table "recruitment_status" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "name" {
    type = varchar(150)
    null = false
  }
  column "hex_code" {
    type = varchar(10)
    null = true
  }
  column "is_active" {
    type    = boolean
    default = true
    null    = false
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

  index "idx_recruitment_status_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "provinsi" {
  schema = schema.public

  column "id" {
    type = varchar(2)
  }
  column "name" {
    type = varchar(255)
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_provinsi_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "kabupaten_kota" {
  schema = schema.public

  column "id" {
    type = varchar(5)
  }
  column "name" {
    type = varchar(255)
    null = false
  }
  column "provinsi_id" {
    type = varchar(2)
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_kabupaten_kota_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_kabupaten_kota_provinsi_id" {
    columns = [column.provinsi_id]
  }
}