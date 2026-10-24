# Changelog

All notable changes to **sdk-go** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0/).

## [0.3.0] - 2026-10-24

### Added

- Typed clients for every catalog product (Todo, Desk, Scheduler, Pulse, Postern, Invest, Abiso, Bisita, Tala, Aklat, Flow, Perfecto, Verdict, Storage, PolicyManager, Hanap, SandboxPayments, Bantay, AiCenter, DatasetBuilder).
- Catalog host overrides for policy-manager, dataset-builder, and sandbox-payments. AI Center health uses `/healthz`.
- Yaman `Secrets().List`, `Update`, and `Delete`.

## [0.2.0] - 2026-10-23

### Added

- Typed Yaman client: `Secrets().Get(name)` and `Secrets().Create({ name, value, description? })`. Default host `https://yaman.raspberry-vision.cloud`.

## [0.1.0] - 2026-10-23

### Added

- Platform HTTP client with Bearer / `X-Api-Key` auth (JWT skips `X-Api-Key`).
- Default Raspberry Vision hosts and generic `Service(slug)` requests.
- Typed Tracker client: events, deployments, `NotifyDeploy`, systems, completeness ingest.
