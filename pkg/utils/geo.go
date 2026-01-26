package utils

import "github.com/umahmood/haversine"

// DistanceMeters berechnet die Entfernung in Metern zwischen zwei geografischen Punkten
func DistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	pos1 := haversine.Coord{Lat: lat1, Lon: lon1}
	pos2 := haversine.Coord{Lat: lat2, Lon: lon2}

	_, km := haversine.Distance(pos1, pos2)
	return km * 1000
}

// IsWithinRadius prüft ob ein Punkt innerhalb eines Radius liegt
func IsWithinRadius(lat1, lon1, lat2, lon2 float64, radiusMeters float64) bool {
	return DistanceMeters(lat1, lon1, lat2, lon2) <= radiusMeters
}
