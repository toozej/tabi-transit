import { useLocalSearchParams } from "expo-router";
import { RouteView } from "@/features/riderInfo/RiderInfoViews";
export default function RouteRoute() {
  const { routeId } = useLocalSearchParams<{ routeId: string }>();
  return <RouteView id={routeId ?? ""} />;
}
