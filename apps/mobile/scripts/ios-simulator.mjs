#!/usr/bin/env node

import { readFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import path from "node:path";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const mobileDirectory = path.resolve(scriptDirectory, "..");
const configPath = path.join(mobileDirectory, "ios-simulators.json");
const config = JSON.parse(readFileSync(configPath, "utf8"));

function fail(message) {
  console.error(`ios-simulator: ${message}`);
  process.exit(1);
}

function execute(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd ?? mobileDirectory,
    encoding: "utf8",
    stdio: options.inherit ? "inherit" : "pipe",
  });

  if (result.error) {
    fail(`could not run ${command}: ${result.error.message}`);
  }
  if (result.status !== 0) {
    const detail = result.stderr?.trim() || result.stdout?.trim();
    fail(`${command} ${args.join(" ")} failed${detail ? `: ${detail}` : ""}`);
  }

  return result.stdout?.trim() ?? "";
}

function compareVersions(left, right) {
  const leftParts = left.split(".").map(Number);
  const rightParts = right.split(".").map(Number);
  const length = Math.max(leftParts.length, rightParts.length);

  for (let index = 0; index < length; index += 1) {
    const difference = (leftParts[index] ?? 0) - (rightParts[index] ?? 0);
    if (difference !== 0) {
      return difference;
    }
  }

  return 0;
}

function targetFor(targetId) {
  const target = config.targets.find((candidate) => candidate.id === targetId);
  if (!target) {
    fail(
      `unknown target ${targetId}; choose ${config.targets.map(({ id }) => id).join(", ")}`,
    );
  }
  return target;
}

function requireMacHost() {
  if (process.platform !== "darwin") {
    fail("CoreSimulator is only available on macOS");
  }

  const macOSVersion = execute("sw_vers", ["-productVersion"]);
  if (compareVersions(macOSVersion, "15.6") < 0) {
    fail(
      `macOS ${macOSVersion} is unsupported; ${config.host.macOS} is required`,
    );
  }

  const xcodeVersionOutput = execute("xcodebuild", ["-version"]);
  const xcodeVersion = xcodeVersionOutput.match(/^Xcode (\S+)/m)?.[1];
  if (!xcodeVersion) {
    fail("could not determine the selected Xcode version");
  }
  if (compareVersions(xcodeVersion, config.host.xcode) !== 0) {
    console.error(
      `ios-simulator: selected Xcode ${xcodeVersion}; this Sequoia configuration is pinned to Xcode ${config.host.xcode}`,
    );
  }
}

function simctlList(category) {
  const output = execute("xcrun", ["simctl", "list", "-j", category]);
  try {
    return JSON.parse(output);
  } catch (error) {
    fail(`could not parse simctl ${category} output: ${error.message}`);
  }
}

function availableRuntimes(target) {
  const runtimes = simctlList("runtimes").runtimes.filter(
    (runtime) =>
      runtime.isAvailable &&
      runtime.platform === "iOS" &&
      Number(runtime.version.split(".")[0]) === target.runtimeMajor,
  );
  runtimes.sort((left, right) => compareVersions(right.version, left.version));
  return runtimes;
}

function latestRuntime(target) {
  const runtimes = availableRuntimes(target);

  if (runtimes.length === 0) {
    fail(
      `no available iOS ${target.runtimeMajor}.x runtime is installed. On the Intel Mac, install the latest listed ${target.runtimeMajor}.x runtime in Xcode > Settings > Components, or run: xcodebuild -downloadPlatform iOS -buildVersion ${target.recommendedRuntime} -architectureVariant ${config.host.runtimeArchitecture}`,
    );
  }

  return runtimes[0];
}

function installRuntimePrerequisite(target) {
  const installed = availableRuntimes(target).find(
    ({ version }) => compareVersions(version, target.recommendedRuntime) >= 0,
  );
  if (installed) {
    console.error(
      `ios-simulator: iOS ${installed.version} satisfies the ${target.runtimeMajor}.x runtime requirement`,
    );
    return;
  }

  console.error(
    `ios-simulator: installing iOS ${target.recommendedRuntime} universal Simulator runtime`,
  );
  execute(
    "xcodebuild",
    [
      "-downloadPlatform",
      "iOS",
      "-buildVersion",
      target.recommendedRuntime,
      "-architectureVariant",
      config.host.runtimeArchitecture,
    ],
    { inherit: true },
  );

  const downloaded = availableRuntimes(target).find(
    ({ version }) => compareVersions(version, target.recommendedRuntime) >= 0,
  );
  if (!downloaded) {
    fail(
      `Xcode completed the download but no available iOS ${target.runtimeMajor}.x runtime was found`,
    );
  }
}

function installXcodePrerequisites() {
  console.error("ios-simulator: installing required Xcode system components");
  execute("xcodebuild", ["-runFirstLaunch"], { inherit: true });
  for (const target of config.targets) {
    installRuntimePrerequisite(target);
  }
}

function validateDeviceType(target) {
  const deviceTypes = simctlList("devicetypes").devicetypes;
  if (
    !deviceTypes.some(
      ({ identifier }) => identifier === target.deviceTypeIdentifier,
    )
  ) {
    fail(
      `${target.deviceTypeIdentifier} is not available in the selected Xcode installation`,
    );
  }
}

function ensureSimulator(targetId) {
  const target = targetFor(targetId);
  const runtime = latestRuntime(target);
  validateDeviceType(target);

  const simulatorName = `${target.name} (iOS ${runtime.version})`;
  const devices = simctlList("devices").devices[runtime.identifier] ?? [];
  const existing = devices.find(
    (device) => device.isAvailable && device.name === simulatorName,
  );

  if (existing) {
    console.error(`ios-simulator: using ${simulatorName} (${existing.udid})`);
    return { target, runtime, name: simulatorName, udid: existing.udid };
  }

  const udid = execute("xcrun", [
    "simctl",
    "create",
    simulatorName,
    target.deviceTypeIdentifier,
    runtime.identifier,
  ]);
  console.error(`ios-simulator: created ${simulatorName} (${udid})`);
  return { target, runtime, name: simulatorName, udid };
}

function bootSimulator(targetId) {
  const simulator = ensureSimulator(targetId);
  const devices =
    simctlList("devices").devices[simulator.runtime.identifier] ?? [];
  const state = devices.find(({ udid }) => udid === simulator.udid)?.state;

  if (state !== "Booted") {
    execute("xcrun", ["simctl", "boot", simulator.udid]);
  }
  execute("xcrun", ["simctl", "bootstatus", simulator.udid, "-b"], {
    inherit: true,
  });
  execute("open", [
    "-a",
    "Simulator",
    "--args",
    "-CurrentDeviceUDID",
    simulator.udid,
  ]);

  return simulator;
}

function printConfiguration() {
  console.log(
    `Host: ${config.host.macOS}; Xcode ${config.host.xcode}; ${config.host.runtimeArchitecture} runtimes`,
  );
  for (const target of config.targets) {
    console.log(
      `${target.id}: ${target.name}, latest installed iOS ${target.runtimeMajor}.x (download baseline ${target.recommendedRuntime})`,
    );
  }
}

const [action = "list", targetId] = process.argv.slice(2);

if (action === "list") {
  printConfiguration();
  process.exit(0);
}

requireMacHost();

if (action === "prereqs" || action === "bootstrap") {
  installXcodePrerequisites();
} else if (action === "ensure-all") {
  for (const target of config.targets) {
    ensureSimulator(target.id);
  }
} else if (action === "ensure") {
  if (!targetId) {
    fail("ensure requires a target ID");
  }
  console.log(ensureSimulator(targetId).udid);
} else if (action === "boot") {
  if (!targetId) {
    fail("boot requires a target ID");
  }
  console.log(bootSimulator(targetId).udid);
} else if (action === "run") {
  if (!targetId) {
    fail("run requires a target ID");
  }
  const simulator = bootSimulator(targetId);
  execute(
    "corepack",
    ["pnpm", "exec", "expo", "run:ios", "--device", simulator.udid],
    { inherit: true },
  );
} else {
  fail(
    "expected list, prereqs, ensure-all, ensure <target>, boot <target>, or run <target>",
  );
}
