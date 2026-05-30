schema "public" {
}

enum "user_account_status" {
  schema = schema.public
  values = ["Active", "Inactive", "Blocked"]
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
  column "summary" {
    type = text
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
  column "system_role_id" {
    type = uuid
    null = true
  }
  column "assigned_role_id" {
    type = uuid
    null = true
  }
  column "account_status" {
    type = enum.user_account_status
    null = true
  }
  column "job_title_id" {
    type = uuid
    null = true
  }
  column "verified_at" {
    type = timestamptz
    null = true
  }
  column "must_reset_password" {
    type    = boolean
    default = false
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



  foreign_key "user_system_role_fk" {
    columns     = [column.system_role_id]
    ref_columns = [table.system_roles.column.id]
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

  foreign_key "user_assigned_role_fk" {
    columns     = [column.assigned_role_id]
    ref_columns = [table.job_roles.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }

  foreign_key "user_job_title_fk" {
    columns     = [column.job_title_id]
    ref_columns = [table.job_titles.column.id]
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
  
  index "idx_users_assigned_role_id" {
    columns = [column.assigned_role_id]
  }

  index "idx_users_email" {
    columns = [column.email]
    unique  = true
  }
}

table "user_tokens" {
  schema = schema.public

  column "id" {
    type    = uuid
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "type" {
    type = varchar(50)
    null = false
  }
  column "token" {
    type = varchar(255)
    null = false
  }
  column "expires_at" {
    type = timestamptz
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }

  primary_key {
    columns = [column.id]
  }

  foreign_key "user_tokens_user_id_fk" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  index "idx_user_tokens_token" {
    columns = [column.token]
    unique  = true
  }

  index "idx_user_tokens_user_id_type" {
    columns = [column.user_id, column.type]
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

table "system_roles" {
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

  index "idx_system_roles_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "job_roles" {
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

  index "idx_job_roles_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_job_roles_sector_id" {
    columns = [column.sector_id]
  }

  foreign_key "job_roles_sector_fk" {
    columns     = [column.sector_id]
    ref_columns = [table.sectors.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "job_titles" {
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

  index "idx_job_titles_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_job_titles_sector_id" {
    columns = [column.sector_id]
  }

  foreign_key "job_titles_sector_fk" {
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
  column "can_be_deleted" {
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

table "user_has_job_roles" {
  schema = schema.public

  column "user_id" {
    type = uuid
    null = false
  }
  column "job_role_id" {
    type = uuid
    null = false
  }

  primary_key {
    columns = [column.user_id, column.job_role_id]
  }

  foreign_key "user_has_job_roles_user_fk" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "user_has_job_roles_job_role_fk" {
    columns     = [column.job_role_id]
    ref_columns = [table.job_roles.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
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

table "bookmarks" {
  schema = schema.public

  column "admin_id" {
    type = uuid
    null = false
  }
  column "candidate_id" {
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

  primary_key {
    columns = [column.admin_id, column.candidate_id]
  }

  foreign_key "bookmarks_admin_fk" {
    columns     = [column.admin_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "bookmarks_candidate_fk" {
    columns     = [column.candidate_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

enum "request_urgency" {
  schema = schema.public
  values = ["HIGH", "MIDDLE", "LOW"]
}

enum "request_status" {
  schema = schema.public
  values = ["PENDING", "ACCEPTED", "PROCESSING", "REJECTED", "DONE"]
}

enum "interview_status" {
  schema = schema.public
  values = ["SCHEDULED", "CANCELLED", "RESCHEDULED", "NO_SHOW", "COMPLETED"]
}

enum "interview_method" {
  schema = schema.public
  values = ["Online", "Offline"]
}

enum "subrequest_level" {
  schema = schema.public
  values = ["Junior", "Middle", "Senior"]
}

table "requests" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "project_name" {
    type = varchar(255)
    null = false
  }
  column "due_date" {
    type = date
    null = true
  }
  column "admin_user_id" {
    type = uuid
    null = true
  }
  column "employee_user_id" {
    type = uuid
    null = false
  }
  column "required_headcount" {
    type = int
    null = false
  }
  column "status" {
    type    = enum.request_status
    default = "'PENDING'"
    null    = false
  }
  column "urgency" {
    type = enum.request_urgency
    null = true
  }
  column "fulfillment_date" {
    type = date
    null = true
  }
  column "rejected_reason" {
    type = text
    null = true
  }
  column "project_duration" {
    type = varchar(255)
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

  index "idx_requests_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_requests_admin_user_id" {
    columns = [column.admin_user_id]
  }

  index "idx_requests_employee_user_id" {
    columns = [column.employee_user_id]
  }

  foreign_key "requests_admin_user_fk" {
    columns     = [column.admin_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "requests_employee_user_fk" {
    columns     = [column.employee_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "subrequests" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "request_id" {
    type = uuid
    null = false
  }
  column "job_role_id" {
    type = uuid
    null = true
  }
  column "tech_stack" {
    type = jsonb
    null = true
  }
  column "notes" {
    type = text
    null = true
  }
  column "is_filled" {
    type    = boolean
    default = false
    null    = false
  }
  column "overview" {
    type = text
    null = true
  }
  column "level" {
    type = enum.subrequest_level
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

  index "idx_subrequests_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_subrequests_request_id" {
    columns = [column.request_id]
  }

  index "idx_subrequests_job_role_id" {
    columns = [column.job_role_id]
  }

  foreign_key "subrequests_request_fk" {
    columns     = [column.request_id]
    ref_columns = [table.requests.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "subrequests_job_role_fk" {
    columns     = [column.job_role_id]
    ref_columns = [table.job_roles.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }
}

table "interview_stages" {
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
  column "hex_code" {
    type = varchar(10)
    null = true
  }
  column "is_active" {
    type    = boolean
    default = true
    null    = false
  }

  primary_key {
    columns = [column.id]
  }

  index "idx_interview_stages_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "interviews" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "admin_user_id" {
    type = uuid
    null = true
  }
  column "candidate_user_id" {
    type = uuid
    null = false
  }
  column "subrequest_id" {
    type = uuid
    null = false
  }
  column "stage_id" {
    type = uuid
    null = false
  }
  column "title" {
    type = varchar(255)
    null = true
  }
  column "description" {
    type = text
    null = true
  }
  column "scheduled_at" {
    type = timestamptz
    null = true
  }
  column "method" {
    type = enum.interview_method
    null = true
  }
  column "status" {
    type = enum.interview_status
    null = true
  }
  column "meeting_link" {
    type = text
    null = true
  }
  column "meeting_location" {
    type = text
    null = true
  }
  column "is_email_sent" {
    type    = boolean
    default = false
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

  index "idx_interviews_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_interviews_admin_user_id" {
    columns = [column.admin_user_id]
  }

  index "idx_interviews_candidate_user_id" {
    columns = [column.candidate_user_id]
  }

  index "idx_interviews_subrequest_id" {
    columns = [column.subrequest_id]
  }

  index "idx_interviews_stage_id" {
    columns = [column.stage_id]
  }

  foreign_key "interviews_candidate_user_fk" {
    columns     = [column.candidate_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "interviews_admin_user_fk" {
    columns     = [column.admin_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }

  foreign_key "interviews_subrequest_fk" {
    columns     = [column.subrequest_id]
    ref_columns = [table.subrequests.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "interviews_stage_fk" {
    columns     = [column.stage_id]
    ref_columns = [table.interview_stages.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }
}

enum "message_attachment_type" {
  schema = schema.public
  values = ["INTERVIEW", "OFFERING"]
}

table "conversations" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "admin_user_id" {
    type = uuid
    null = false
  }
  column "candidate_user_id" {
    type = uuid
    null = false
  }
  column "subrequest_id" {
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

  index "idx_conversations_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_conversations_admin_user_id" {
    columns = [column.admin_user_id]
  }

  index "idx_conversations_candidate_user_id" {
    columns = [column.candidate_user_id]
  }

  foreign_key "conversations_admin_user_fk" {
    columns     = [column.admin_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "conversations_candidate_user_fk" {
    columns     = [column.candidate_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "conversations_subrequest_fk" {
    columns     = [column.subrequest_id]
    ref_columns = [table.subrequests.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  index "idx_conversations_subrequest_id" {
    columns = [column.subrequest_id]
  }
}

table "messages" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "conversation_id" {
    type = uuid
    null = false
  }
  column "content" {
    type = text
    null = false
  }
  column "sender_user_id" {
    type = uuid
    null = false
  }
  column "reply_to_message_id" {
    type = uuid
    null = true
  }
  column "read_at" {
    type = timestamptz
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

  index "idx_messages_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_messages_conversation_id" {
    columns = [column.conversation_id]
  }

  index "idx_messages_sender_user_id" {
    columns = [column.sender_user_id]
  }

  index "idx_messages_reply_to_message_id" {
    columns = [column.reply_to_message_id]
  }

  foreign_key "messages_conversation_fk" {
    columns     = [column.conversation_id]
    ref_columns = [table.conversations.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "messages_sender_user_fk" {
    columns     = [column.sender_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "messages_reply_to_message_fk" {
    columns     = [column.reply_to_message_id]
    ref_columns = [table.messages.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }
}

table "offerings" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "filename" {
    type = varchar(255)
    null = false
  }
  column "user_id_kandidat" {
    type = uuid
    null = false
  }
  column "path" {
    type = varchar(255)
    null = false
  }
  column "subrequest_id" {
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

  index "idx_offerings_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_offerings_user_id_kandidat" {
    columns = [column.user_id_kandidat]
  }

  index "idx_offerings_subrequest_id" {
    columns = [column.subrequest_id]
  }

  foreign_key "offerings_user_kandidat_fk" {
    columns     = [column.user_id_kandidat]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "offerings_subrequest_fk" {
    columns     = [column.subrequest_id]
    ref_columns = [table.subrequests.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "message_attachments" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "message_id" {
    type = uuid
    null = false
  }
  column "attachment_type" {
    type = enum.message_attachment_type
    null = false
  }
  column "interview_id" {
    type = uuid
    null = true
  }
  column "offering_id" {
    type = uuid
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

  index "idx_message_attachments_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_message_attachments_message_id" {
    columns = [column.message_id]
  }

  foreign_key "message_attachments_message_fk" {
    columns     = [column.message_id]
    ref_columns = [table.messages.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "message_attachments_interview_fk" {
    columns     = [column.interview_id]
    ref_columns = [table.interviews.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }

  foreign_key "message_attachments_offering_fk" {
    columns     = [column.offering_id]
    ref_columns = [table.offerings.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }
}

table "onboard_histories" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "candidate_user_id" {
    type = uuid
    null = false
  }
  column "offering_id" {
    type = uuid
    null = true
  }
  column "start_date" {
    type = date
    null = true
  }
  column "end_date" {
    type = date
    null = true
  }
  column "snapshot" {
    type = jsonb
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

  index "idx_onboard_histories_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_onboard_histories_candidate_user_id" {
    columns = [column.candidate_user_id]
  }

  index "idx_onboard_histories_offering_id" {
    columns = [column.offering_id]
  }

  foreign_key "onboard_histories_candidate_user_fk" {
    columns     = [column.candidate_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "onboard_histories_offering_fk" {
    columns     = [column.offering_id]
    ref_columns = [table.offerings.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

enum "review_final_recommendation" {
  schema = schema.public
  values = ["HIGHLY_RECOMMENDED", "RECOMMENDED", "CONSIDERED", "NOT_RECOMMENDED"]
}

table "reviews" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "subrequest_id" {
    type = uuid
    null = false
  }
  column "candidate_user_id" {
    type = uuid
    null = false
  }
  column "employee_user_id" {
    type = uuid
    null = false
  }
  column "onboard_history_id" {
    type = uuid
    null = false
  }
  column "work_quality" {
    type = int
    null = true
  }
  column "timeliness" {
    type = int
    null = true
  }
  column "communication_collaboration" {
    type = int
    null = true
  }
  column "problem_solving_initiative" {
    type = int
    null = true
  }
  column "final_recommendation" {
    type = enum.review_final_recommendation
    null = true
  }
  column "notes" {
    type = text
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

  index "idx_reviews_deleted_at" {
    columns = [column.deleted_at]
  }

  index "idx_reviews_subrequest_id" {
    columns = [column.subrequest_id]
  }

  index "idx_reviews_candidate_user_id" {
    columns = [column.candidate_user_id]
  }

  index "idx_reviews_employee_user_id" {
    columns = [column.employee_user_id]
  }

  index "idx_reviews_onboard_history_id" {
    columns = [column.onboard_history_id]
  }

  foreign_key "reviews_subrequest_fk" {
    columns     = [column.subrequest_id]
    ref_columns = [table.subrequests.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "reviews_candidate_user_fk" {
    columns     = [column.candidate_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "reviews_employee_user_fk" {
    columns     = [column.employee_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "reviews_onboard_history_fk" {
    columns     = [column.onboard_history_id]
    ref_columns = [table.onboard_histories.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "system_settings" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "is_ai_mode_enabled" {
    type    = boolean
    default = false
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

  index "idx_system_settings_deleted_at" {
    columns = [column.deleted_at]
  }
}

enum "job_vacancy_schema" {
  schema = schema.public
  values = ["ONSITE", "REMOTE", "HYBRID"]
}

enum "job_vacancy_status" {
  schema = schema.public
  values = ["DRAFT", "ARCHIVED", "PUBLISHED"]
}

enum "approval_request_table_name" {
  schema = schema.public
  values = ["faqs", "job_vacancies", "company_profiles", "career_departments"]
}

enum "approval_request_action" {
  schema = schema.public
  values = ["CREATE", "UPDATE", "DELETE"]
}

enum "approval_request_status" {
  schema = schema.public
  values = ["PENDING", "APPROVED", "REJECTED"]
}

table "company_profiles" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "address" {
    type = text
    null = true
  }
  column "phone" {
    type = varchar(20)
    null = true
  }
  column "email" {
    type = varchar(255)
    null = true
  }
  column "facebook_url" {
    type = text
    null = true
  }
  column "instagram_url" {
    type = text
    null = true
  }
  column "linkedin_url" {
    type = text
    null = true
  }
  column "twitter_url" {
    type = text
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

  index "idx_company_profiles_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "career_departments" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "name" {
    type = varchar(50)
    null = true
  }
  column "description" {
    type = text
    null = true
  }
  column "image_path" {
    type = text
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

  index "idx_career_departments_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "faqs" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "question" {
    type = text
    null = false
  }
  column "answer" {
    type = text
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

  index "idx_faqs_deleted_at" {
    columns = [column.deleted_at]
  }
}

table "job_vacancies" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "subrequest_id" {
    type = uuid
    null = false
  }
  column "name" {
    type = varchar(255)
    null = false
  }
  column "takedown_date" {
    type = date
    null = true
  }
  column "fulfillment_date" {
    type = date
    null = true
  }
  column "schema" {
    type = enum.job_vacancy_schema
    null = true
  }
  column "status" {
    type = enum.job_vacancy_status
    null = true
  }
  column "description" {
    type = jsonb
    null = true
  }
  column "candidate_qualification" {
    type = jsonb
    null = true
  }
  column "benefit" {
    type = jsonb
    null = true
  }
  column "overview" {
    type = text
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

  index "idx_job_vacancies_deleted_at" {
    columns = [column.deleted_at]
  }

  foreign_key "job_vacancies_subrequest_fk" {
    columns     = [column.subrequest_id]
    ref_columns = [table.subrequests.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "subrequest_candidates" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "subrequest_id" {
    type = uuid
    null = false
  }
  column "candidate_user_id" {
    type = uuid
    null = false
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

  index "idx_subrequest_candidates_deleted_at" {
    columns = [column.deleted_at]
  }

  foreign_key "subrequest_candidates_subrequest_fk" {
    columns     = [column.subrequest_id]
    ref_columns = [table.subrequests.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "subrequest_candidates_user_fk" {
    columns     = [column.candidate_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "log_activities" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "actor_id" {
    type = uuid
    null = true
  }
  column "action_type" {
    type = varchar(100)
    null = false
  }
  column "module" {
    type = varchar(100)
    null = false
  }
  column "description" {
    type = text
    null = true
  }
  column "metadata" {
    type = jsonb
    null = true
  }
  column "ip_address" {
    type = varchar(45)
    null = true
  }
  column "is_success" {
    type = boolean
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

  index "idx_log_activities_deleted_at" {
    columns = [column.deleted_at]
  }

  foreign_key "log_activities_actor_fk" {
    columns     = [column.actor_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "notifications" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "admin_user_id_owner" {
    type = uuid
    null = false
  }
  column "title" {
    type = varchar(255)
    null = false
  }
  column "description" {
    type = text
    null = true
  }
  column "is_read" {
    type = boolean
    default = false
    null = false
  }
  column "is_admin_broadcast" {
    type = boolean
    default = false
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

  index "idx_notifications_deleted_at" {
    columns = [column.deleted_at]
  }

  foreign_key "notifications_admin_fk" {
    columns     = [column.admin_user_id_owner]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "admin_notes" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "admin_user_id" {
    type = uuid
    null = false
  }
  column "candidate_user_id" {
    type = uuid
    null = false
  }
  column "content" {
    type = text
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

  index "idx_admin_notes_deleted_at" {
    columns = [column.deleted_at]
  }

  foreign_key "admin_notes_admin_fk" {
    columns     = [column.admin_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "admin_notes_candidate_fk" {
    columns     = [column.candidate_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "approval_requests" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "requested_by_admin_id" {
    type = uuid
    null = false
  }
  column "table_name" {
    type = enum.approval_request_table_name
    null = false
  }
  column "record_id" {
    type = uuid
    null = false
  }
  column "action" {
    type = enum.approval_request_action
    null = false
  }
  column "proposed_data" {
    type = jsonb
    null = true
  }
  column "status" {
    type = enum.approval_request_status
    default = "'PENDING'"
    null = false
  }
  column "rejected_reason" {
    type = text
    null = true
  }
  column "reviewed_by_superadmin_id" {
    type = uuid
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

  index "idx_approval_requests_deleted_at" {
    columns = [column.deleted_at]
  }

  foreign_key "approval_req_admin_fk" {
    columns     = [column.requested_by_admin_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }

  foreign_key "approval_req_superadmin_fk" {
    columns     = [column.reviewed_by_superadmin_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }
}

table "ai_chats" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "admin_user_id" {
    type    = uuid
    null    = false
  }
  column "title" {
    type = varchar(255)
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

  index "idx_ai_chats_deleted_at" {
    columns = [column.deleted_at]
  }

  foreign_key "ai_chats_admin_fk" {
    columns     = [column.admin_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "ai_messages" {
  schema = schema.public

  column "id" {
    type = uuid
  }
  column "ai_chat_id" {
    type = uuid
    null = false
  }
  column "role" {
    type = varchar(20)
    null = false
  }
  column "content" {
    type = text
    null = false
  }
  column "is_last_message" {
    type = boolean
    default = false
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

  index "idx_ai_messages_deleted_at" {
    columns = [column.deleted_at]
  }

  foreign_key "ai_messages_chat_fk" {
    columns     = [column.ai_chat_id]
    ref_columns = [table.ai_chats.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}