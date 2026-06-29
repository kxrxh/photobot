interface PageLoadingProps {
	message?: string
}

export function PageLoading({ message = "Загрузка…" }: PageLoadingProps) {
	return (
		<div className="flex justify-center items-center min-h-[400px]">
			<div className="flex flex-col items-center gap-4">
				<span className="loading loading-spinner loading-lg text-primary" />
				<p className="text-base-content/70">{message}</p>
			</div>
		</div>
	)
}
