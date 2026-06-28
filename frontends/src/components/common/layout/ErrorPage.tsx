import { HTTPError } from "ky"
import type { FC } from "react"
import { FaExclamationTriangle } from "react-icons/fa"
import { TELEGRAM_VALIDATION_ERROR_MESSAGE } from "@/lib/auth/messages"
import { getUserFacingErrorMessage } from "@/utils/errors"

type ErrorPageProps = {
	error: Error
	fullHeight?: boolean
}

const ErrorPage: FC<ErrorPageProps> = ({ error, fullHeight = true }) => {
	const errorMessage =
		error instanceof HTTPError && error.response.status === 401
			? TELEGRAM_VALIDATION_ERROR_MESSAGE
			: getUserFacingErrorMessage(error)

	const containerHeightClass = fullHeight ? "min-h-screen" : "min-h-[240px] w-full"

	return (
		<div className={`flex items-center justify-center p-4 ${containerHeightClass}`}>
			<div className="text-center">
				<FaExclamationTriangle className="mx-auto mb-4 text-4xl text-primary" />
				<h2 className="mb-4 text-3xl font-semibold text-primary">Что-то пошло не так :(</h2>
				<p className="mb-8 text-base-content/70">{errorMessage}</p>
			</div>
		</div>
	)
}

export default ErrorPage
