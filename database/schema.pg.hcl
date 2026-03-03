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



  foreign_key "user_role_system_fk" {
    columns     = [column.role_system_id]
    ref_columns = [table.role_systems.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }

  foreign_key "user_kabupaten_kota_fk" {
    columns     = [column.kabupaten_kota_id]
    ref_columns = [table.kabupaten_kota.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }

  foreign_key "user_recruitment_status_fk" {
    columns     = [column.recruitment_status_id]
    ref_columns = [table.recruitment_statuses.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }

  foreign_key "user_role_applied_fk" {
    columns     = [column.role_applied_id]
    ref_columns = [table.role_applieds.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
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

  index "idx_users_email" {
    columns = [column.email]
    unique  = true
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

  foreign_key "cv_user_fk" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
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

table "role_systems" {
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

  index "idx_role_systems_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "role_applieds" {
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

  index "idx_role_applieds_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_role_applieds_sector_id" {
    columns = [column.sector_id]
  }

  foreign_key "role_applieds_sector_fk" {
    columns     = [column.sector_id]
    ref_columns = [table.sectors.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "recruitment_statuses" {
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

  index "idx_recruitment_statuses_deleted_at" {
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

  foreign_key "kabupaten_kota_provinsi_id_fk" {
    columns     = [column.provinsi_id]
    ref_columns = [table.provinsi.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}