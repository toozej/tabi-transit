import { useMutation, useQuery } from "@tanstack/react-query";
import { lazy, type FormEvent, Suspense, useEffect, useState } from "react";
import {
  Link,
  useNavigate,
  useParams,
  useSearchParams,
} from "react-router-dom";

import { validOpaqueId } from "../app/urls";
import { webConfig } from "../app/config";
import {
  getAlerts,
  getArrivals,
  getNearby,
  getRoute,
  getRouteStops,
  getStop,
  getVehicles,
  planJourney,
  queryKeys,
  type NearbyPosition,
} from "./api";
import { freshnessLabel, type Vehicle } from "./models";
import { createBrowserRepository, type SavedItem } from "../platform/storage";

const VehicleMap = lazy(async () => import("../maps/VehicleMap"));

function AsyncState({
  loading,
  error,
  children,
}: {
  loading: boolean;
  error: boolean;
  children: React.ReactNode;
}) {
  if (loading) return <p aria-live="polite">Loading rider information.</p>;
  if (error)
    return (
      <p role="alert">
        Rider information is unavailable. Cached or fixture data is not
        presented as live.
      </p>
    );
  return <>{children}</>;
}

export function NearbyPage() {
  const [position, setPosition] = useState<NearbyPosition>();
  const [locationDenied, setLocationDenied] = useState(false);
  const query = useQuery({
    queryKey: [...queryKeys.nearby, position] as const,
    queryFn: () => getNearby(position),
    enabled:
      position !== undefined ||
      import.meta.env.VITE_TABI_DATA_MODE !== "remote",
  });
  return (
    <section className="page">
      <h1>Nearby stops</h1>
      <p>
        Choose location access to find nearby stops, or select a stop below.
      </p>
      <button
        type="button"
        onClick={() => {
          if (!navigator.geolocation) {
            setLocationDenied(true);
            return;
          }
          navigator.geolocation.getCurrentPosition(
            (current) => {
              setLocationDenied(false);
              setPosition({
                latitude: current.coords.latitude,
                longitude: current.coords.longitude,
              });
            },
            () => setLocationDenied(true),
          );
        }}
      >
        Use my location
      </button>
      {locationDenied && (
        <p role="alert">
          Location access was not granted. You can still open a stop from Saved
          items or a shared link.
        </p>
      )}
      <AsyncState loading={query.isLoading} error={query.isError}>
        <ul>
          {query.data?.map((stop) => (
            <li key={stop.id}>
              <Link to={`/stops/${stop.id}`}>{stop.name}</Link>
              {stop.distanceMeters ? ` · ${stop.distanceMeters} m away` : ""}
            </li>
          ))}
        </ul>
      </AsyncState>
    </section>
  );
}

function VehicleList({
  vehicles,
  onVehicleSelect,
}: {
  vehicles: readonly Vehicle[];
  onVehicleSelect: (id: string) => void;
}) {
  return (
    <ul className="vehicle-list">
      {vehicles.map((vehicle) => (
        <li key={vehicle.id}>
          <button
            className="vehicle-select"
            type="button"
            onClick={() => onVehicleSelect(vehicle.id)}
          >
            Vehicle {vehicle.sourceVehicleId}
          </button>
          <span>
            {vehicle.routeId ?? "Route unavailable"} ·{" "}
            {freshnessLabel(vehicle.freshness)}
          </span>
          <Link to={`/vehicles/${vehicle.id}`}>Open vehicle details</Link>
        </li>
      ))}
    </ul>
  );
}

export function MapPage() {
  const [search, setSearch] = useState("");
  const [params, setParams] = useSearchParams();
  const query = useQuery({
    queryKey: queryKeys.vehicles,
    queryFn: getVehicles,
  });
  const matching =
    query.data?.filter(
      (vehicle) => !search || vehicle.sourceVehicleId === search.trim(),
    ) ?? [];
  const selectedVehicleId = validOpaqueId(params.get("vehicle") ?? undefined);
  const selected = query.data?.find(
    (vehicle) => vehicle.id === selectedVehicleId,
  );
  const selectVehicle = (id: string) => {
    const next = new URLSearchParams(params);
    next.set("vehicle", id);
    setParams(next);
  };
  const mapboxAccessToken = webConfig.mapboxAccessToken;
  return (
    <section className="page">
      <h1>Vehicles</h1>
      {mapboxAccessToken ? (
        <Suspense fallback={<p aria-live="polite">Loading vehicle map.</p>}>
          <VehicleMap
            accessToken={mapboxAccessToken}
            styleUrl={webConfig.mapboxStyleUrl}
            vehicles={matching}
            selectedVehicleId={selectedVehicleId}
            onVehicleSelect={selectVehicle}
          />
        </Suspense>
      ) : (
        <p className="notice">
          Vehicle map rendering needs the restricted public Mapbox browser
          token. Every vehicle task remains available in this accessible list.
        </p>
      )}
      <label>
        Search exact vehicle ID{" "}
        <input
          value={search}
          onChange={(event) => setSearch(event.target.value)}
        />
      </label>
      <AsyncState loading={query.isLoading} error={query.isError}>
        {selected && (
          <section className="vehicle-selection" aria-live="polite">
            <h2>Vehicle {selected.sourceVehicleId}</h2>
            <p>
              {selected.routeId ?? "Route unavailable"} ·{" "}
              {freshnessLabel(selected.freshness)}
            </p>
            <Link to={`/vehicles/${selected.id}`}>Open vehicle details</Link>
          </section>
        )}
        <VehicleList vehicles={matching} onVehicleSelect={selectVehicle} />
      </AsyncState>
    </section>
  );
}

export function VehiclePage() {
  const id = validOpaqueId(useParams().vehicleId);
  const query = useQuery({
    queryKey: queryKeys.vehicles,
    queryFn: getVehicles,
  });
  const vehicle = query.data?.find((item) => item.id === id);
  if (!id) return <NotFound />;
  return (
    <section className="page">
      <AsyncState loading={query.isLoading} error={query.isError}>
        {vehicle ? (
          <>
            <h1>Vehicle {vehicle.sourceVehicleId}</h1>
            <dl>
              <dt>Route</dt>
              <dd>{vehicle.routeId ?? "Unavailable"}</dd>
              <dt>Destination</dt>
              <dd>{vehicle.headsign ?? "Unavailable"}</dd>
              <dt>Freshness</dt>
              <dd>{freshnessLabel(vehicle.freshness)}</dd>
            </dl>
            {vehicle.freshness.status !== "fresh" && (
              <p role="alert">This vehicle is not shown as live.</p>
            )}
            <h2>History</h2>
            <p>
              History is unavailable in fixture preview. It is never presented
              as live location.
            </p>
          </>
        ) : (
          <NotFound />
        )}
      </AsyncState>
    </section>
  );
}

export function AlertsPage() {
  const query = useQuery({ queryKey: queryKeys.alerts, queryFn: getAlerts });
  return (
    <section className="page">
      <h1>Alerts</h1>
      <AsyncState loading={query.isLoading} error={query.isError}>
        <ul>
          {query.data?.map((alert) => (
            <li key={alert.id}>
              <Link to={`/alerts/${alert.id}`}>{alert.header}</Link>
              <p>{alert.description}</p>
              <small>{freshnessLabel(alert.freshness)}</small>
            </li>
          ))}
        </ul>
      </AsyncState>
    </section>
  );
}

export function AlertPage() {
  const id = validOpaqueId(useParams().alertId);
  return id ? <AlertsPage /> : <NotFound />;
}
export function StopPage() {
  const id = validOpaqueId(useParams().stopId);
  const stop = useQuery({
    queryKey: ["stop", id],
    queryFn: () => getStop(id!),
    enabled: id !== undefined,
  });
  const arrivals = useQuery({
    queryKey: ["arrivals", id],
    queryFn: () => getArrivals(id!),
    enabled: id !== undefined,
  });
  return id ? (
    <section className="page">
      <AsyncState loading={stop.isLoading} error={stop.isError}>
        <h1>{stop.data?.name ?? "Stop"}</h1>
        <p>Stop ID: {id}</p>
        <h2>Arrivals</h2>
        <AsyncState loading={arrivals.isLoading} error={arrivals.isError}>
          {arrivals.data?.length ? (
            <ul>
              {arrivals.data.map((arrival) => (
                <li key={arrival.id}>
                  {arrival.routeId} to{" "}
                  {arrival.headsign ?? "destination unavailable"}:{" "}
                  {arrival.status}
                  {arrival.estimatedAt ? " (estimated)" : " (scheduled)"}
                </li>
              ))}
            </ul>
          ) : (
            <p>Arrivals unavailable; no realtime claim is shown.</p>
          )}
        </AsyncState>
        <h2>Offline schedule</h2>
        <p>
          Schedule is unavailable offline until a static-feed artifact is
          downloaded.
        </p>
      </AsyncState>
    </section>
  ) : (
    <NotFound />
  );
}
export function RoutePage() {
  const id = validOpaqueId(useParams().routeId);
  const route = useQuery({
    queryKey: ["route", id],
    queryFn: () => getRoute(id!),
    enabled: id !== undefined,
  });
  const routeStops = useQuery({
    queryKey: ["route-stops", id],
    queryFn: () => getRouteStops(id!),
    enabled: id !== undefined,
  });
  return id ? (
    <section className="page">
      <AsyncState loading={route.isLoading} error={route.isError}>
        <h1>
          {route.data
            ? `${route.data.route.shortName} ${route.data.route.longName}`
            : "Route"}
        </h1>
        <p>Route ID: {id}</p>
        <h2>Stops</h2>
        <AsyncState loading={routeStops.isLoading} error={routeStops.isError}>
          {routeStops.data?.length ? (
            <ol>
              {routeStops.data.map((stop) => (
                <li key={stop.id}>
                  <Link to={`/stops/${stop.id}`}>{stop.name}</Link>
                </li>
              ))}
            </ol>
          ) : (
            <p>Route stop sequence is unavailable.</p>
          )}
        </AsyncState>
      </AsyncState>
    </section>
  ) : (
    <NotFound />
  );
}

export function PlanPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const [origin, setOrigin] = useState(params.get("origin") ?? "");
  const [destination, setDestination] = useState(
    params.get("destination") ?? "",
  );
  const plan = useMutation({
    mutationFn: () => planJourney(origin, destination),
  });
  function submit(event: FormEvent) {
    event.preventDefault();
    const query = new URLSearchParams();
    if (validOpaqueId(origin)) query.set("origin", origin);
    if (validOpaqueId(destination)) query.set("destination", destination);
    navigate(`/plan?${query}`);
    if (validOpaqueId(origin) && validOpaqueId(destination)) plan.mutate();
  }
  return (
    <section className="page">
      <h1>Plan a trip</h1>
      <form onSubmit={submit}>
        <label>
          Origin ID{" "}
          <input
            value={origin}
            onChange={(event) => setOrigin(event.target.value)}
            placeholder="Saved or stop ID"
          />
        </label>
        <label>
          Destination ID{" "}
          <input
            value={destination}
            onChange={(event) => setDestination(event.target.value)}
            placeholder="Saved or stop ID"
          />
        </label>
        <button>Plan trip</button>
      </form>
      <p>
        Share links contain opaque place IDs only. Search and geocoding remain
        behind the Tabi API.
      </p>
      {plan.isError && <p role="alert">{plan.error.message}</p>}
      {plan.data && (
        <section>
          <h2>Itineraries</h2>
          <p>
            Ranking is supplied by the Tabi API; provider payloads are not
            shown.
          </p>
          {plan.data.itineraries.length ? (
            <ol>
              {plan.data.itineraries.map((itinerary) => (
                <li key={itinerary.id}>
                  {Math.round(itinerary.durationSeconds / 60)} minutes ·{" "}
                  {itinerary.transfers} transfers · {itinerary.walkingMeters} m
                  walking
                </li>
              ))}
            </ol>
          ) : (
            <p>No itinerary meets these constraints.</p>
          )}
        </section>
      )}
    </section>
  );
}

export function SavedPage() {
  const [items, setItems] = useState<{
    saved: SavedItem[];
    recents: SavedItem[];
  }>({ saved: [], recents: [] });
  const repository = createBrowserRepository();
  useEffect(() => {
    void repository.read().then(setItems);
  }, []);
  return (
    <section className="page">
      <h1>Saved</h1>
      <p>
        Browser data is device-local. If storage is unavailable, changes are
        session-only.
      </p>
      <h2>Saved items</h2>
      {items.saved.length ? (
        <ul>
          {items.saved.map((item) => (
            <li key={item.id}>{item.label}</li>
          ))}
        </ul>
      ) : (
        <p>No saved stops, routes, or vehicles yet.</p>
      )}
      <h2>Recent</h2>
      {items.recents.length ? (
        <ul>
          {items.recents.map((item) => (
            <li key={item.id}>{item.label}</li>
          ))}
        </ul>
      ) : (
        <p>No recent selections.</p>
      )}
      <button
        onClick={() =>
          void repository
            .clear("recents")
            .then(() => setItems((current) => ({ ...current, recents: [] })))
        }
      >
        Clear recent selections
      </button>
      <button
        onClick={() =>
          void repository
            .clear("all")
            .then(() => setItems({ saved: [], recents: [] }))
        }
      >
        Clear saved data
      </button>
    </section>
  );
}

export function SettingsPage() {
  return (
    <section className="page">
      <h1>Settings</h1>
      <h2>Notifications</h2>
      <p>
        Browser push is not enabled. It requires separate consent and
        service-worker approval.
      </p>
      <h2>Privacy</h2>
      <p>
        Tabi does not store precise location, search text, access tokens, or
        notification secrets in browser storage by default.
      </p>
    </section>
  );
}
export function CreditsPage() {
  return (
    <section className="page">
      <h1>Credits and attribution</h1>
      <p>
        Transit information is supplied through the Tabi API. Map rendering is
        currently unavailable in the web client.
      </p>
    </section>
  );
}
export function NotFound() {
  return (
    <section className="page">
      <h1>Page not found</h1>
      <Link to="/nearby">Go to nearby stops</Link>
    </section>
  );
}
