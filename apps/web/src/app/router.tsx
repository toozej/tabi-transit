import {
  Navigate,
  createBrowserRouter,
  RouterProvider,
} from "react-router-dom";
import { Shell } from "../components/Shell";
import {
  AlertPage,
  AlertsPage,
  CreditsPage,
  MapPage,
  NearbyPage,
  NotFound,
  PlanPage,
  RoutePage,
  SavedPage,
  SettingsPage,
  StopPage,
  VehiclePage,
} from "../features/Pages";

const router = createBrowserRouter([
  {
    element: <Shell />,
    children: [
      { index: true, element: <Navigate to="/nearby" replace /> },
      { path: "/nearby", element: <NearbyPage /> },
      { path: "/map", element: <MapPage /> },
      { path: "/vehicles/:vehicleId", element: <VehiclePage /> },
      { path: "/plan", element: <PlanPage /> },
      { path: "/alerts", element: <AlertsPage /> },
      { path: "/alerts/:alertId", element: <AlertPage /> },
      { path: "/saved", element: <SavedPage /> },
      { path: "/stops/:stopId", element: <StopPage /> },
      { path: "/routes/:routeId", element: <RoutePage /> },
      { path: "/settings/*", element: <SettingsPage /> },
      { path: "/credits", element: <CreditsPage /> },
      { path: "*", element: <NotFound /> },
    ],
  },
]);
export function Router() {
  return <RouterProvider router={router} />;
}
