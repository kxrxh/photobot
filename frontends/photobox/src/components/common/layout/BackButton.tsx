import { useNavigate } from "@tanstack/react-router"
import { FaArrowLeft } from "react-icons/fa"

export function BackButton() {
	const navigate = useNavigate()

	return (
		<button
			type="button"
			className="w-full btn btn-primary"
			onClick={() => navigate({ to: "/menu" })}
		>
			<FaArrowLeft className="mr-2" />
			Назад в меню
		</button>
	)
}
