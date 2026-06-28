import type React from "react"
import type { Analysis, KalibriObject } from "@/api/analysis/types"

export interface TabComponentProps {
	analysis: Analysis
	objects: KalibriObject[] | undefined
	objectsLoading: boolean
	objectsError: boolean
}

export type TabType = "main" | "characteristics" | "objects"

export interface TabDefinition {
	id: TabType
	label: string
	icon: React.ComponentType<{ size?: number }>
}
