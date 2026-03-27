env "dev" {
  src = "file://schema"
  dev = "docker://postgres/17/dev?search_path=public"
  url = getenv("DATABASE_URL")

  migration {
    dir              = "file://migrations"
    revisions_schema = "public"
  }

  # Exclude the Atlas revision table from schema diffs so it doesn't
  # appear as "drift" when comparing desired vs actual state.
  exclude = ["atlas_schema_revisions"]
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
