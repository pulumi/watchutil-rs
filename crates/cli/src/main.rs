use std::path::{PathBuf};

use clap::{Parser, StructOpt};

/// Monitor a directory and emit a newline line on stdout on change.
#[derive(Parser, Debug, Clone)]
#[structopt(name = "pulumi-watch")]
pub struct ProgramArgs {
  #[structopt(required = true, value_parser, value_name = "DIR", value_hint = clap::ValueHint::DirPath)]
  path: PathBuf,
}

#[derive(StructOpt, Debug, Clone)]
struct Options {}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
  Ok(())
}
