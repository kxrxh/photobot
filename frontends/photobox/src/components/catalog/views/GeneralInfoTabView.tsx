import type { WeedImage } from "@/api/catalog/types"
import Gallery from "@/components/common/ui/Gallery"

export interface GeneralInfoTabViewProps {
	images: WeedImage[]
	description: string | null
	mainGroup?: string
	mainSubgroup?: string
	subgroup?: string
	isQuarantine: boolean
	harmfulness?: string
}

export default function GeneralInfoTabView({
	images,
	description,
	mainGroup,
	mainSubgroup,
	subgroup,
	isQuarantine,
	harmfulness,
}: GeneralInfoTabViewProps) {
	return (
		<div className="p-4 space-y-6">
			<Gallery
				images={images.map((image) => image.url)}
				altText="Фото образца"
				className="w-full"
			/>
			{mainGroup && (
				<div>
					<h3 className="font-semibold text-base-content/70">Классификация</h3>
					<p>{mainGroup}</p>
				</div>
			)}
			{mainSubgroup && (
				<div>
					<p>{mainSubgroup}</p>
				</div>
			)}
			{subgroup && (
				<div>
					<p>{subgroup}</p>
				</div>
			)}
			{description && (
				<div>
					<h3 className="font-semibold text-base-content/70">Описание</h3>
					<p>{description}</p>
				</div>
			)}
			{harmfulness && (
				<div>
					<h3 className="font-semibold text-base-content/70">Вредоносность</h3>
					<p>{harmfulness}</p>
				</div>
			)}
			<div>
				<h3 className="font-semibold text-base-content/70">Карантинное</h3>
				<p>{isQuarantine ? "Да" : "Нет"}</p>
			</div>
		</div>
	)
}
