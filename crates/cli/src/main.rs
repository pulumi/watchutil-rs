use std::{path::PathBuf, sync::Arc, time::Duration, collections::HashSet};

use clap::{Parser, StructOpt};
use miette::{Diagnostic, IntoDiagnostic, Result};
use thiserror::Error;
use watchexec::{
  action::{Action, Outcome},
  config::{InitConfig, RuntimeConfig},
  handler::PrintDebug,
  signal::source::MainSignal,
  Watchexec,
};
use watchexec_filterer_globset::GlobsetFilterer;

/// Monitor a directory and emit a newline line on stdout on change.
#[derive(Parser, Debug, Clone)]
#[structopt(name = "pulumi-watch")]
pub struct ProgramArgs {
  // Root directory, or project origin, to watch. Used to configure filters.
  #[structopt(short = 'o', long, required = true, value_parser, value_name = "DIR", value_hint = clap::ValueHint::DirPath)]
  origin: PathBuf,

  // Directory to watch. May be specified multiple times.
  #[structopt(short = 'w', long, required = true, value_parser, value_name = "DIR", value_hint = clap::ValueHint::DirPath)]
  watch: Vec<PathBuf>,
}

#[derive(StructOpt, Debug, Clone)]
struct Options {}

#[tokio::main]
async fn main() -> Result<()> {
  let args = ProgramArgs::parse();

  let mut init = InitConfig::default();
  init.on_error(PrintDebug(std::io::stderr()));

  use std::path::MAIN_SEPARATOR as SEP;
  let ignores: Vec<(String, Option<PathBuf>)> = vec![
    // Default ignores used by watchexec.
    (format!("**{s}.DS_Store", s = SEP), None),
    (String::from("*.py[co]"), None),
    (String::from("#*#"), None),
    (String::from(".#*"), None),
    (String::from(".*.kate-swp"), None),
    (String::from(".*.sw?"), None),
    (String::from(".*.sw?x"), None),
    (format!("**{s}.bzr{s}**", s = SEP), None),
    (format!("**{s}_darcs{s}**", s = SEP), None),
    (format!("**{s}.fossil-settings{s}**", s = SEP), None),
    (format!("**{s}.git{s}**", s = SEP), None),
    (format!("**{s}.hg{s}**", s = SEP), None),
    (format!("**{s}.pijul{s}**", s = SEP), None),
    (format!("**{s}.svn{s}**", s = SEP), None),
  ];

  let mut runtime = RuntimeConfig::default();
  runtime.pathset(args.watch);
  runtime.action_throttle(Duration::from_millis(250));

  // Note: to preserve existing watch behavior, we do not read/interpret other ignore files.
  let filter = GlobsetFilterer::new(args.origin, [], ignores, [], [])
    .await
    .into_diagnostic()?;

  runtime.filterer(Arc::new(filter));
  let we = Watchexec::new(init, runtime.clone()).into_diagnostic()?;

  runtime.on_action(move |action: Action| async move {
    let mut paths = HashSet::new();
    for p in action.events.clone().iter() {
      for (path, _) in p.paths() {
        paths.insert(path.display().to_string());
      }
      for s in p.signals() {
        if let MainSignal::Interrupt | MainSignal::Quit | MainSignal::Terminate = s {
          action.outcome(Outcome::Exit);
          return Ok(());
        }
      }
    }
    if !paths.is_empty() {
      eprintln!("pulumi-watch: event on paths ${paths:?}");
    }

    let timestamp = chrono::Utc::now();
    let timestamp = timestamp.format("%+");
    println!("{timestamp}");

    Ok(()) as Result<(), NoneError>
  });

  we.reconfigure(runtime).into_diagnostic()?;
  we.main().await.into_diagnostic()?.into_diagnostic()?;
  Ok(())
}

/// Only needed because our handler is infallible - it just emits to stdout.
#[derive(Debug, Error, Diagnostic)]
#[error("stub")]
struct NoneError;
