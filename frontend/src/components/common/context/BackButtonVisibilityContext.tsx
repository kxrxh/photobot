import { createContext, useContext } from "react"

interface BackButtonVisibilityContextType {
	isBackButtonHidden: boolean
	setIsBackButtonHidden: (isHidden: boolean) => void
}

export const BackButtonVisibilityContext = createContext<
	BackButtonVisibilityContextType | undefined
>(undefined)

export const useBackButtonVisibility = () => {
	const context = useContext(BackButtonVisibilityContext)
	if (context === undefined) {
		throw new Error("useBackButtonVisibility must be used within a BackButtonVisibilityProvider")
	}
	return context
}
