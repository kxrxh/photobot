import { useQueryClient } from "@tanstack/react-query"
import { AnimatePresence, motion } from "framer-motion"
import { useEffect, useState } from "react"
import type { Classification, Fraction } from "@/api/classification/types"
import { queryKeys } from "@/api/queryKeys"
import ClassificationConstructor from "@/components/classification/components/ClassificationConstructor"
import ClassificationList from "@/components/classification/components/ClassificationList"
import { useBackButtonVisibility } from "@/components/common/context/BackButtonVisibilityContext"
import { PageType } from "@/routes/_authenticated/classification"

export default function ClassificationPage() {
	const queryClient = useQueryClient()
	const [pageState, setPageState] = useState<{
		page: PageType
		classification?: Classification
		isCopyMode?: boolean
		onSaveSuccess?: () => void
		fractions?: Fraction[]
	}>({ page: PageType.LIST })

	const setCurrentPage = (
		page: PageType,
		classification?: Classification,
		isCopyMode?: boolean,
		options?: { onSaveSuccess?: () => void; fractions?: Fraction[] }
	) => {
		const newState = { page, classification, isCopyMode, ...(options || {}) }
		setPageState(newState)

		if (page === "constructor" || page === "copy") {
			window.history.pushState({ page, classification, isCopyMode }, "", window.location.href)
		}
	}

	const { setIsBackButtonHidden } = useBackButtonVisibility()

	useEffect(() => {
		const handlePopState = (event: PopStateEvent) => {
			if (event.state && (event.state.page === "constructor" || event.state.page === "copy")) {
				setPageState({
					page: event.state.page === "copy" ? PageType.COPY : PageType.CONSTRUCTOR,
					classification: event.state.classification,
					isCopyMode: event.state.isCopyMode,
				})
			} else {
				setPageState({ page: PageType.LIST })
			}
		}

		window.addEventListener("popstate", handlePopState)
		return () => window.removeEventListener("popstate", handlePopState)
	}, [])

	useEffect(() => {
		if (!window.history.state && pageState.page === PageType.LIST) {
			window.history.replaceState({ page: PageType.LIST }, "", window.location.href)
		}
	}, [pageState.page])

	useEffect(() => {
		if (pageState.page === PageType.CONSTRUCTOR || pageState.page === PageType.COPY) {
			setIsBackButtonHidden(true)
		} else {
			setIsBackButtonHidden(false)
		}
	}, [pageState.page, setIsBackButtonHidden])

	const variants = {
		initial: { opacity: 0 },
		animate: { opacity: 1 },
		exit: { opacity: 0 },
	}

	return (
		<AnimatePresence mode="wait">
			<motion.div
				key={pageState.page}
				variants={variants}
				initial="initial"
				animate="animate"
				exit="exit"
				transition={{ duration: 0.2 }}
				className="w-full"
			>
				{(() => {
					switch (pageState.page) {
						case PageType.LIST:
							return (
								<ClassificationList
									setCurrentPage={(...args) => {
										const [page, classification, isCopyMode, options] = args
										if (page === PageType.COPY) {
											const onSaveSuccess = () => {
												queryClient.invalidateQueries({
													queryKey: queryKeys.classifications.all,
												})
												if (options?.onSaveSuccess) options.onSaveSuccess()
												setPageState({ page: PageType.LIST })
											}
											setCurrentPage(page, classification, isCopyMode, {
												...options,
												onSaveSuccess,
											})
										} else {
											setCurrentPage(...args)
										}
									}}
								/>
							)
						case PageType.CONSTRUCTOR:
							return (
								<ClassificationConstructor
									initialData={pageState.classification}
									isCopyMode={pageState.isCopyMode}
									onSaveSuccess={pageState.onSaveSuccess}
								/>
							)
						case PageType.COPY:
							return (
								<ClassificationConstructor
									initialData={pageState.classification}
									isCopyMode={true}
									onSaveSuccess={pageState.onSaveSuccess}
									fractionsData={pageState.fractions}
								/>
							)
						default:
							return null
					}
				})()}
			</motion.div>
		</AnimatePresence>
	)
}
