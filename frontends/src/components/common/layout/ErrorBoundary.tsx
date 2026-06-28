import type { ReactNode } from "react"
import { Component as ReactComponent } from "react"
import { FaExclamationTriangle } from "react-icons/fa"
import { getUserFacingErrorMessage } from "@/utils/errors"

interface Props {
	children: ReactNode
	fallback?: ReactNode
}

interface State {
	hasError: boolean
	error: Error | null
}

/**
 * Error boundary for catching runtime errors in the component tree.
 * Renders a fallback UI when an error is thrown.
 */
export class ErrorBoundary extends ReactComponent<Props, State> {
	constructor(props: Props) {
		super(props)
		this.state = { hasError: false, error: null }
	}

	static getDerivedStateFromError(error: Error): State {
		return { hasError: true, error }
	}

	override render() {
		if (this.state.hasError && this.state.error) {
			if (this.props.fallback) {
				return this.props.fallback
			}
			const message = getUserFacingErrorMessage(this.state.error)
			return (
				<div className="flex min-h-[240px] w-full items-center justify-center p-4">
					<div className="text-center">
						<FaExclamationTriangle className="mx-auto mb-4 text-4xl text-primary" />
						<h2 className="mb-4 text-3xl font-semibold text-primary">Что-то пошло не так :(</h2>
						<p className="mb-8 text-base-content/70">{message}</p>
					</div>
				</div>
			)
		}
		return this.props.children
	}
}
