import { useEffect, useMemo, useState } from "react"
import { AttributionControl, MapContainer, Marker, TileLayer, useMapEvents } from "react-leaflet"
import "leaflet/dist/leaflet.css"
import L from "leaflet"
import icon from "leaflet/dist/images/marker-icon.png"
import iconShadow from "leaflet/dist/images/marker-shadow.png"
import { FaCheck, FaMapMarkerAlt } from "react-icons/fa"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import { log } from "@/utils/log"

const DefaultIcon = L.icon({
	iconUrl: icon,
	shadowUrl: iconShadow,
	iconSize: [25, 41],
	iconAnchor: [12, 41],
})
L.Marker.prototype.options.icon = DefaultIcon
interface LocationMapSheetProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: (coords: string) => void
	initialLocation?: string
}

// Helper component to handle clicks on the map and view movement
function MapController({
	position,
	setPosition,
	isOpen,
}: {
	position: L.LatLng | null
	setPosition: (pos: L.LatLng) => void
	isOpen: boolean
}) {
	const map = useMapEvents({
		click(e) {
			setPosition(e.latlng)
		},
	})

	// Force invalidate size when component mounts
	useEffect(() => {
		if (!isOpen) return
		const tid = window.setTimeout(() => {
			map.invalidateSize()
		}, 100)
		return () => clearTimeout(tid)
	}, [map, isOpen])

	// Move view when position changes (e.g., from geolocation)
	useEffect(() => {
		if (position) {
			map.setView(position, map.getZoom())
		}
	}, [position, map])

	return position === null ? null : <Marker position={position} />
}

export default function LocationMapSheet({
	isOpen,
	onClose,
	onConfirm,
	initialLocation,
}: LocationMapSheetProps) {
	const [position, setPosition] = useState<L.LatLng | null>(null)

	// Parse initial location if it exists
	useEffect(() => {
		if (isOpen && initialLocation) {
			try {
				const [lat, lng] = initialLocation.split(",").map((s) => parseFloat(s.trim()))
				if (!Number.isNaN(lat) && !Number.isNaN(lng)) {
					setPosition(new L.LatLng(lat, lng))
				}
			} catch (e) {
				log.devError("Failed to parse initial location", e)
			}
		}
	}, [isOpen, initialLocation])

	// Try to center the map on the device; user can still tap the map to pick a point.
	useEffect(() => {
		if (!isOpen || position !== null) return
		if (!navigator.geolocation) {
			setPosition((current) => current ?? new L.LatLng(55.75, 37.61))
			return
		}

		navigator.geolocation.getCurrentPosition(
			(pos) => {
				setPosition(new L.LatLng(pos.coords.latitude, pos.coords.longitude))
			},
			(error) => {
				// Code 1 = user blocked permission — expected, not an app bug.
				if (error.code !== error.PERMISSION_DENIED) {
					log.devWarn("Geolocation failed:", error.message)
				}
				setPosition((current) => current ?? new L.LatLng(55.75, 37.61))
			},
			{ enableHighAccuracy: true, timeout: 5000, maximumAge: 0 }
		)
	}, [isOpen, position])

	const mapContent = useMemo(() => {
		if (!isOpen) return null

		// Use a fixed center if position isn't loaded yet to avoid MapContainer issues
		const center = position || new L.LatLng(55.75, 37.61)

		return (
			<MapContainer
				center={center}
				zoom={13}
				style={{ height: "100%", width: "100%", minHeight: "400px" }}
				attributionControl={false}
			>
				<AttributionControl prefix='<a href="https://leafletjs.com" title="A JavaScript library for interactive maps">Leaflet</a>' />
				<TileLayer url="https://{s}.tile.openstreetmap.fr/hot/{z}/{x}/{y}.png" maxZoom={19} />

				<MapController position={position} setPosition={setPosition} isOpen={isOpen} />
			</MapContainer>
		)
	}, [isOpen, position])

	if (!isOpen) return null

	return (
		<div className="fixed inset-0 z-100 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm">
			<div className="bg-base-100 w-full max-w-3xl rounded-xl shadow-2xl flex flex-col max-h-[90vh]">
				<div className="p-4 border-b border-base-200 flex justify-between items-center">
					<h3 className="font-bold text-lg flex items-center gap-2">
						<FaMapMarkerAlt className="text-primary" />
						Выбрать местоположение
					</h3>
					<SheetHeaderCloseButton onClick={onClose} aria-label="Закрыть карту" />
				</div>

				<div className="flex-1 relative min-h-[400px] bg-base-200">
					{mapContent}

					<div className="absolute top-4 left-1/2 -translate-x-1/2 z-400 bg-base-100/90 px-4 py-2 rounded-full shadow-lg text-xs font-medium backdrop-blur-md border border-base-200 pointer-events-none">
						Нажмите на карту чтобы поставить метку
					</div>
				</div>

				<div className="p-4 border-t border-base-200 flex justify-end gap-2 bg-base-100 rounded-b-xl">
					<button type="button" onClick={onClose} className="btn flex-1">
						Отмена
					</button>
					<button
						type="button"
						onClick={() => {
							if (position) {
								onConfirm(`${position.lat.toFixed(5)}, ${position.lng.toFixed(5)}`)
								onClose()
							}
						}}
						className="btn btn-primary flex-1"
						disabled={!position}
					>
						<FaCheck className="mr-2" />
						Подтвердить
					</button>
				</div>
			</div>
		</div>
	)
}
