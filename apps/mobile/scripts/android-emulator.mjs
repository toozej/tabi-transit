#!/usr/bin/env node

import {
  accessSync,
  constants,
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
} from "node:fs";
import { homedir } from "node:os";
import { spawn, spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import path from "node:path";

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const mobileDirectory = path.resolve(scriptDirectory, "..");
const configPath = path.join(mobileDirectory, "android-emulators.json");
const brewfilePath = path.join(mobileDirectory, "Brewfile.android");
const config = JSON.parse(readFileSync(configPath, "utf8"));
const sdkRoot =
  process.env.ANDROID_SDK_ROOT ??
  process.env.ANDROID_HOME ??
  path.join(homedir(), "Library", "Android", "sdk");
const avdHome =
  process.env.ANDROID_AVD_HOME ?? path.join(homedir(), ".android", "avd");

function fail(message) {
  console.error(`android-emulator: ${message}`);
  process.exit(1);
}

function execute(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd ?? mobileDirectory,
    encoding: "utf8",
    env: options.env ?? process.env,
    input: options.input,
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

function commandPath(command, candidates = []) {
  for (const candidate of candidates) {
    try {
      accessSync(candidate, constants.X_OK);
      return candidate;
    } catch {
      // Try the next explicit location before consulting PATH.
    }
  }

  const result = spawnSync("which", [command], { encoding: "utf8" });
  if (result.status === 0 && result.stdout.trim()) {
    return result.stdout.trim();
  }
  fail(`${command} is unavailable; run make prereqs-android-simulators`);
}

function androidEnvironment() {
  const environment = {
    ...process.env,
    ANDROID_HOME: sdkRoot,
    ANDROID_SDK_ROOT: sdkRoot,
    ANDROID_AVD_HOME: avdHome,
  };
  const java21 = "/usr/local/opt/openjdk@21/libexec/openjdk.jdk/Contents/Home";
  const bundledJava =
    "/Applications/Android Studio.app/Contents/jbr/Contents/Home";
  if (!environment.JAVA_HOME && existsSync(java21)) {
    environment.JAVA_HOME = java21;
  } else if (!environment.JAVA_HOME && existsSync(bundledJava)) {
    environment.JAVA_HOME = bundledJava;
  }
  return environment;
}

function requireIntelMac() {
  if (process.platform !== "darwin") {
    fail("the configured Android Studio host is macOS");
  }
  if (process.arch !== "x64") {
    fail(
      `the configured ${config.host.systemImageAbi} system images require the Intel Mac host`,
    );
  }

  const macOSVersion = execute("sw_vers", ["-productVersion"]);
  const [major, minor] = macOSVersion.split(".").map(Number);
  if (major < 15 || (major === 15 && minor < 6)) {
    fail(
      `macOS ${macOSVersion} is unsupported; ${config.host.macOS} is required`,
    );
  }
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

function sdkTool(command) {
  return commandPath(command, [
    path.join(sdkRoot, "cmdline-tools", "latest", "bin", command),
    path.join(sdkRoot, "platform-tools", command),
    path.join(sdkRoot, "emulator", command),
  ]);
}

function installPrerequisites() {
  const brew = commandPath("brew");
  execute(brew, ["bundle", "install", `--file=${brewfilePath}`], {
    inherit: true,
  });

  const brewPrefix = execute(brew, ["--prefix"]);
  const sdkmanager = commandPath("sdkmanager", [
    path.join(brewPrefix, "bin", "sdkmanager"),
    path.join(
      brewPrefix,
      "share",
      "android-commandlinetools",
      "cmdline-tools",
      "latest",
      "bin",
      "sdkmanager",
    ),
  ]);
  const environment = androidEnvironment();
  mkdirSync(sdkRoot, { recursive: true });
  mkdirSync(avdHome, { recursive: true });

  console.error(
    "android-emulator: review and accept the Android SDK licenses when prompted",
  );
  execute(sdkmanager, [`--sdk_root=${sdkRoot}`, "--licenses"], {
    env: environment,
    inherit: true,
  });

  const packages = [
    ...config.sdkPackages,
    ...config.targets.map(({ systemImage }) => systemImage),
  ];
  execute(sdkmanager, [`--sdk_root=${sdkRoot}`, ...new Set(packages)], {
    env: environment,
    inherit: true,
  });

  execute(sdkTool("emulator"), ["-accel-check"], {
    env: environment,
    inherit: true,
  });
}

function deviceDefinitions(avdmanager) {
  const output = execute(avdmanager, ["list", "device"], {
    env: androidEnvironment(),
  });
  return output.split(/^-{5,}$/m).flatMap((block) => {
    const id = block.match(/^id:\s+\d+\s+or\s+"([^"]+)"/m)?.[1];
    const name = block.match(/^\s*Name:\s+(.+)$/m)?.[1]?.trim();
    return id && name ? [{ id, name }] : [];
  });
}

function baseDeviceFor(target, avdmanager) {
  const definitions = deviceDefinitions(avdmanager);
  for (const candidate of target.deviceCandidates) {
    const normalized = candidate.toLowerCase();
    const match = definitions.find(
      ({ id, name }) =>
        id.toLowerCase() === normalized || name.toLowerCase() === normalized,
    );
    if (match) {
      return match;
    }
  }
  fail(
    `none of the base profiles for ${target.displayName} are installed: ${target.deviceCandidates.join(", ")}`,
  );
}

function installedAvds(avdmanager) {
  const output = execute(avdmanager, ["list", "avd"], {
    env: androidEnvironment(),
  });
  return output.split(/^-{5,}$/m).flatMap((block) => {
    const name = block.match(/^\s*Name:\s+(.+)$/m)?.[1]?.trim();
    const avdPath = block.match(/^\s*Path:\s+(.+)$/m)?.[1]?.trim();
    return name && avdPath ? [{ name, path: avdPath }] : [];
  });
}

function updateAvdConfiguration(configFile, target) {
  const values = new Map();
  for (const line of readFileSync(configFile, "utf8").split("\n")) {
    const separator = line.indexOf("=");
    if (separator > 0) {
      values.set(line.slice(0, separator), line.slice(separator + 1));
    }
  }

  const sharedHardware = {
    "disk.dataPartition.size": "8G",
    "fastboot.forceColdBoot": "no",
    "fastboot.forceFastBoot": "yes",
    "hw.gpu.enabled": "yes",
    "hw.gpu.mode": "auto",
    "hw.keyboard": "yes",
    "hw.mainKeys": "no",
    "hw.sdCard": "yes",
    showDeviceFrame: "no",
    "vm.heapSize": "512",
  };
  for (const [key, value] of Object.entries({
    ...sharedHardware,
    ...target.hardware,
  })) {
    values.set(key, value);
  }

  const content = [...values]
    .map(([key, value]) => `${key}=${value}`)
    .join("\n");
  writeFileSync(configFile, `${content}\n`);
}

function ensureEmulator(targetId) {
  const target = targetFor(targetId);
  const avdmanager = sdkTool("avdmanager");
  const environment = androidEnvironment();
  const existing = installedAvds(avdmanager).find(
    ({ name }) => name === target.avdName,
  );
  let avdPath = existing?.path;

  if (!avdPath) {
    const baseDevice = baseDeviceFor(target, avdmanager);
    avdPath = path.join(avdHome, `${target.avdName}.avd`);
    execute(
      avdmanager,
      [
        "create",
        "avd",
        "--name",
        target.avdName,
        "--package",
        target.systemImage,
        "--device",
        baseDevice.id,
        "--path",
        avdPath,
      ],
      { env: environment, input: "no\n" },
    );
    console.error(
      `android-emulator: created ${target.displayName} from ${baseDevice.name}`,
    );
  } else {
    console.error(`android-emulator: using ${target.avdName}`);
  }

  const avdConfig = path.join(avdPath, "config.ini");
  if (!existsSync(avdConfig)) {
    fail(`AVD configuration is missing: ${avdConfig}`);
  }
  updateAvdConfiguration(avdConfig, target);
  return target;
}

function runningSerial(target, adb) {
  const result = spawnSync(adb, ["devices"], {
    encoding: "utf8",
    env: androidEnvironment(),
  });
  if (result.status !== 0) {
    return undefined;
  }

  const serials = result.stdout
    .split("\n")
    .slice(1)
    .map((line) => line.trim().split(/\s+/))
    .filter(([, state]) => state === "device")
    .map(([serial]) => serial)
    .filter((serial) => serial.startsWith("emulator-"));

  return serials.find((serial) => {
    const name = spawnSync(adb, ["-s", serial, "emu", "avd", "name"], {
      encoding: "utf8",
      env: androidEnvironment(),
    });
    return (
      name.status === 0 &&
      name.stdout.trim().split(/\r?\n/)[0] === target.avdName
    );
  });
}

function waitForEmulator(target, adb) {
  // A newly created AVD may need to expand its disk image before Android can
  // report boot completion; three minutes is not reliable on a cold host.
  const deadline = Date.now() + 300_000;
  while (Date.now() < deadline) {
    const serial = runningSerial(target, adb);
    if (serial) {
      const booted = spawnSync(
        adb,
        ["-s", serial, "shell", "getprop", "sys.boot_completed"],
        { encoding: "utf8", env: androidEnvironment() },
      );
      if (booted.status === 0 && booted.stdout.trim() === "1") {
        return serial;
      }
    }
    Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 1_000);
  }
  fail(`${target.displayName} did not finish booting within five minutes`);
}

function bootEmulator(targetId) {
  const target = ensureEmulator(targetId);
  const adb = sdkTool("adb");
  const emulator = sdkTool("emulator");
  let serial = runningSerial(target, adb);

  if (!serial) {
    const child = spawn(
      emulator,
      ["-avd", target.avdName, "-no-snapshot-save"],
      { detached: true, env: androidEnvironment(), stdio: "ignore" },
    );
    child.unref();
    console.error(`android-emulator: starting ${target.avdName}`);
    serial = waitForEmulator(target, adb);
  }
  return { target, serial };
}

function printConfiguration() {
  console.log(
    `Host: ${config.host.macOS}; ${config.host.architecture}; ${config.host.systemImageAbi} images`,
  );
  for (const target of config.targets) {
    console.log(
      `${target.id}: ${target.displayName}, Android ${target.androidVersion} (API ${target.apiLevel})`,
    );
  }
}

const [action = "list", targetId] = process.argv.slice(2);

if (action === "list") {
  printConfiguration();
  process.exit(0);
}

requireIntelMac();

if (action === "prereqs") {
  installPrerequisites();
} else if (action === "ensure-all") {
  for (const target of config.targets) {
    ensureEmulator(target.id);
  }
} else if (action === "ensure") {
  if (!targetId) {
    fail("ensure requires a target ID");
  }
  console.log(ensureEmulator(targetId).avdName);
} else if (action === "run") {
  if (!targetId) {
    fail("run requires a target ID");
  }
  const { target } = bootEmulator(targetId);
  execute(
    "corepack",
    ["pnpm", "exec", "expo", "run:android", "--device", target.avdName],
    { env: androidEnvironment(), inherit: true },
  );
} else {
  fail("expected list, prereqs, ensure-all, ensure <target>, or run <target>");
}
