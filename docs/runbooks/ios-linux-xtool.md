# iOS development on Linux: xtool assessment

Assessment date: 2026-10-14. CLI version: xtool 1.22.2. Host: Linux x86_64.

xtool cannot build the current Tabi iOS application. Tabi uses Expo and React
Native with native modules, including Mapbox and Expo SQLite. Expo generates
an Xcode project and uses CocoaPods for native dependencies. xtool builds
Swift Package Manager applications. It does not replace this build process.

The [upstream React Native support issue](https://github.com/xtool-org/xtool/issues/28)
remains open. The maintainer states that xtool can sign applications from
these frameworks, but it cannot build them. Adding a `Package.swift` file
alone would not convert Tabi or its native dependencies to a supported build.

## CLI checks on this host

The assessment used the official Linux AppImage. The binary is in the ignored
`.tools/xtool/1.22.2/` directory. No host GUI was used.

To repeat the download from the repository root:

```bash
mkdir -p .tools/xtool/1.22.2
curl --fail --location \
  https://github.com/xtool-org/xtool/releases/download/1.22.2/xtool-x86_64.AppImage \
  --output .tools/xtool/1.22.2/xtool.AppImage
chmod 0755 .tools/xtool/1.22.2/xtool.AppImage
sha256sum .tools/xtool/1.22.2/xtool.AppImage
```

The downloaded file had this SHA-256 digest:

```text
e3176cdf861f58b7f135c755c7bfaec92d0568ef826027b4e5e3d82b3219ddfe
```

This digest records the tested file. It is not a separate publisher signature.
`APPIMAGE_EXTRACT_AND_RUN=1` permits execution without a FUSE mount.

| Check                                           | Observed result                                                    |
| ----------------------------------------------- | ------------------------------------------------------------------ |
| `xtool --version`                               | Passed: `xtool 1.22.2`                                             |
| `xtool --help`                                  | Passed: CLI starts on Linux                                        |
| `xtool dev build --help`                        | Describes builds for SwiftPM applications in the current directory |
| `xtool new TabiCompatibilityProbe --skip-setup` | Created a SwiftUI package in a temporary directory                 |
| `xtool dev build` in `apps/mobile`              | Failed: `Error: Could not find Package.swift in this directory.`   |
| `expo run:ios --no-install --no-bundler`        | Rejected Linux: iOS builds require macOS                           |

To repeat the build check from the repository root:

```bash
cd apps/mobile
APPIMAGE_EXTRACT_AND_RUN=1 ../../.tools/xtool/1.22.2/xtool.AppImage dev build
```

To repeat the package creation check from the repository root:

```bash
(
  set -euo pipefail
  xtool_cli="$PWD/.tools/xtool/1.22.2/xtool.AppImage"
  scratch=$(mktemp -d)
  trap 'rm -rf "$scratch"' EXIT
  cd "$scratch"
  APPIMAGE_EXTRACT_AND_RUN=1 "$xtool_cli" new TabiCompatibilityProbe --skip-setup
)
```

The package creation command also reported a missing Swift toolchain. This
host has no `swift` command. No Darwin SDK installation, Apple authentication,
signing, device installation, or native execution was attempted. These CLI
checks do not establish a successful native build or an iOS test pass.

## Mobile checks on this host

`make test-js` stopped because Homebrew was unavailable. The assessment then
used the installed repository Node 24.19.0 runtime and dependency executables:

```bash
export PATH="$PWD/.tools/nvm/versions/node/v24.19.0/bin:$PATH"
cd apps/mobile
./node_modules/.bin/vitest run --config vitest.config.mts
../../node_modules/.bin/tsc --noEmit
```

Vitest passed all 59 tests in 20 files. TypeScript reported nine existing
errors in `app.config.ts`, `src/data/api/fixtures.ts`,
`src/domain/vehicleModels.ts`, and `src/test/rntl-proof.test.tsx`. The
assessment changed only documentation. Native execution remains untested.

## Available workflow

Linux can run the JavaScript tests and TypeScript checks. After prerequisite
installation, use `make test-js` and `make typecheck`. These checks do not
execute native Mapbox, SQLite, or iOS UI code.

To remove the Intel Mac from the build process, use a remote macOS build
service, such as EAS Build. Expo documents its
[macOS build infrastructure](https://docs.expo.dev/build-reference/infrastructure/).
The [EAS local build option](https://docs.expo.dev/build-reference/local-builds/)
still requires macOS for iOS builds. Linux containers do not remove that
requirement.

After installation of a compatible development build on an iPhone, Linux can
serve JavaScript through Metro with `make dev-mobile`. The phone must be able
to reach the Linux host. Native dependency changes require another native
build. See Expo's
[development build documentation](https://docs.expo.dev/develop/development-builds/introduction/).

xtool also provides `install` and `launch` commands for an existing application
on a physical device. Those commands do not compile Tabi. Device pairing,
signing, provisioning, and execution require separate validation. xtool does
not provide an iOS simulator runtime for Linux.

The repository still requires the credentials and device evidence described
in [CI and release validation](ci-release.md). This assessment does not enable
a cloud build or change those gates.
