env "dev" {
  src = "file://schema"
  dev = "docker://postgres/17/dev?search_path=public"
  url = getenv("DATABASE_URL")

  migration {
    dir    = "file://migrations"
    format = golang-migrate
  }
}

lint {
  destructive {
    error = true
  }
  data_depend {
    error = true
  }
}

diff {
  skip {
    drop_schema = true
    drop_table  = true
  }
}
