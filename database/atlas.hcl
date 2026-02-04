# declare env
variable "envfile" {
    type    = string
    default = "../.env"
}

# load .env
locals {
    envfile = {
        for line in split("\n", file(var.envfile)): split("=", line)[0] => regex("=(.*)", line)[0]
        if !startswith(line, "#") && length(split("=", line)) > 1
    }
}

# define the data source to load the GORM schema
data "external_schema" "gorm" {
  program = [
    "go", "run", "-mod=mod",
    "ariga.io/atlas-provider-gorm", "load",
    "--path", "./../domain/model/gorm",
    "--dialect", "postgres"
  ]
}

env "hcl" {
    url = local.envfile["POSTGRES_URL"]
    src = "file://database/schema.pg.hcl"
    dev = "docker://postgres/17/atlas_dev"
    migration {
        dir = "file://database/migrations"
    }
    format {
        migrate {
            diff = "{{ sql . \"\" }}"
        }
    }
}
