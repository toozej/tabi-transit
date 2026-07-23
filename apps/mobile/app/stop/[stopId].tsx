import { useLocalSearchParams } from "expo-router";
import { StopView } from "@/features/riderInfo/RiderInfoViews";
export default function StopRoute() {
  const { stopId } = useLocalSearchParams<{ stopId: string }>();
  return <StopView id={stopId ?? ""} />;
}
