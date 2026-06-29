import type React from "react"

interface LoadingSkeletonProps {
	itemCount?: number
}

const LoadingSkeleton: React.FC<LoadingSkeletonProps> = ({ itemCount = 4 }) => (
	<div className="space-y-4">
		{Array.from({ length: itemCount }, (_, index) => `skeleton-${index}`).map((key) => (
			<div
				key={key}
				className="flex items-center p-4 rounded-lg border shadow-sm border-base-300 bg-base-100"
			>
				<div className="flex-grow space-y-2">
					<div className="w-2/3 h-4 skeleton" />
					<div className="w-1/2 h-3 skeleton" />
				</div>
			</div>
		))}
	</div>
)

export default LoadingSkeleton
