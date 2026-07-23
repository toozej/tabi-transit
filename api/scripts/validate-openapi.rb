#!/usr/bin/env ruby
# Minimal dependency-free structural guard. Full semantic linting is performed by Redocly when installed.
require 'yaml'

path = ARGV.fetch(0, File.expand_path('../openapi.yaml', __dir__))
begin
  document = YAML.safe_load_file(path, aliases: false)
rescue Psych::Exception => e
  warn "OpenAPI YAML parse failed: #{e.message}"
  exit 1
end

failures = []
failures << 'openapi must be 3.0.x' unless document['openapi'].to_s.start_with?('3.0.')
%w[info paths components].each { |key| failures << "missing top-level #{key}" unless document.key?(key) }
paths = document['paths'] || {}
%w[/health/live /health/ready /v1/config /v1/routes /v1/stops /v1/stops/nearby /v1/vehicles /v1/vehicles/search /v1/vehicles/{id} /v1/search /v1/geocode/reverse /v1/journeys/plan /v1/installations /v1/installations/{id} /v1/installations/{id}/push-token /v1/subscriptions /v1/subscriptions/{id}].each do |path_name|
  failures << "missing required path #{path_name}" unless paths.key?(path_name)
end
schemas = document.dig('components', 'schemas') || {}
%w[ErrorResponse Freshness ConfigResponse Route Stop Vehicle PlaceResult JourneyPlanRequest JourneyPlanResponse Itinerary InstallationCreateRequest InstallationCreateResponse PushTokenRegistration SubscriptionCreateRequest Subscription SubscriptionCollection].each { |name| failures << "missing schema #{name}" unless schemas.key?(name) }
failures << 'missing installation credential security scheme' unless document.dig('components', 'securitySchemes', 'InstallationCredential')
failures << 'vehicle search must be declared before /v1/vehicles/{id} for unambiguous routing documentation' unless paths.keys.index('/v1/vehicles/search') < paths.keys.index('/v1/vehicles/{id}')

if failures.empty?
  puts "OpenAPI structural validation passed: #{path}"
else
  failures.each { |failure| warn "OpenAPI validation: #{failure}" }
  exit 1
end
