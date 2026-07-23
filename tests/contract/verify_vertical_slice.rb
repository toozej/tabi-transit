#!/usr/bin/env ruby
# Fixture-only contract guard for the first vehicle vertical slice. This is
# intentionally dependency-free so it can run in local and CI environments.
require "json"
require "yaml"

ROOT = File.expand_path("../..", __dir__)
SPEC_PATH = File.join(ROOT, "api/openapi.yaml")

def fail_with(message)
  warn "vertical-slice contract: #{message}"
  exit 1
end

spec = YAML.safe_load_file(SPEC_PATH, aliases: false)
paths = spec.fetch("paths")
%w[/health/live /health/ready /v1/config /v1/vehicles /v1/vehicles/search /v1/vehicles/{id}].each do |path|
  fail_with("missing required endpoint #{path}") unless paths.key?(path)
end

collection = JSON.parse(File.read(File.join(ROOT, "api/examples/vehicle-collection-response.json")))
config = JSON.parse(File.read(File.join(ROOT, "api/examples/config-response.json")))
unavailable = JSON.parse(File.read(File.join(ROOT, "api/examples/source-unavailable-error.json")))

freshness_fields = %w[source fetchedAt processedAt status ageSeconds isRealtime]
valid_statuses = %w[fresh aging stale unknown]
validate_freshness = lambda do |freshness, context|
  fail_with("#{context} lacks freshness") unless freshness.is_a?(Hash)
  freshness_fields.each { |field| fail_with("#{context} freshness missing #{field}") unless freshness.key?(field) }
  fail_with("#{context} has invalid freshness status") unless valid_statuses.include?(freshness.fetch("status"))
  fail_with("#{context} has negative freshness age") unless freshness.fetch("ageSeconds").is_a?(Numeric) && freshness.fetch("ageSeconds") >= 0
end

fail_with("collection lacks a snapshot id") unless collection["snapshotId"].is_a?(String) && !collection["snapshotId"].empty?
validate_freshness.call(collection["freshness"], "collection")
vehicles = collection["vehicles"]
fail_with("collection contains no representative vehicle") unless vehicles.is_a?(Array) && !vehicles.empty?
vehicles.each do |vehicle|
  %w[id sourceVehicleId mode coordinate inService].each { |field| fail_with("vehicle missing #{field}") unless vehicle.key?(field) }
  coordinate = vehicle.fetch("coordinate")
  fail_with("vehicle coordinate is not [longitude, latitude]") unless coordinate.is_a?(Array) && coordinate.length == 2 && coordinate.all? { |value| value.is_a?(Numeric) }
  validate_freshness.call(vehicle["freshness"], "vehicle #{vehicle.fetch("id")}")
end

%w[apiVersion minimumAppVersion features sources staleThresholdSeconds serviceBounds staticFeed].each do |field|
  fail_with("config example missing #{field}") unless config.key?(field)
end
fail_with("vehicle map is not explicitly configured") unless config.dig("features", "vehicleMap", "enabled") == true
fail_with("source-unavailable fixture is not safe") unless unavailable.dig("error", "code") == "source_unavailable" && unavailable.dig("error", "retryAfterSeconds").is_a?(Numeric)

puts "Vertical-slice fixture contract passed (#{vehicles.length} representative vehicle(s))."
