import { FaBox, FaClock, FaInfoCircle } from "react-icons/fa"
import { IoImageOutline } from "react-icons/io5"
import ImageFullscreen from "@/components/common/dialogs/ImageFullscreen"
import { getAnalysisOutputUrls, getAnalysisSourceUrls } from "@/utils/image"
import { translateProductName } from "@/utils/reportTranslations"
import AnalysisSheetSection from "../components/AnalysisSheetSection"
import type { TabComponentProps } from "../types"

const formatDate = (dateStr: string) => {
	const date = new Date(dateStr)
	return new Intl.DateTimeFormat("ru-RU", {
		day: "2-digit",
		month: "2-digit",
		year: "numeric",
		hour: "2-digit",
		minute: "2-digit",
		second: "2-digit",
	}).format(date)
}

const MainTab: React.FC<TabComponentProps> = ({
	analysis,
	objects: _objects,
	objectsLoading: _objectsLoading,
	objectsError: _objectsError,
}) => {
	const imageUrls = getAnalysisSourceUrls(analysis)
	const processedImages = getAnalysisOutputUrls(analysis)

	return (
		<div className="space-y-5 sm:space-y-6">
			<AnalysisSheetSection
				title="Основные данные"
				subtitle="Идентификатор, продукт и время анализа"
				icon={<FaInfoCircle size={18} />}
				accent="primary"
			>
				<div className="space-y-3">
					<div className="flex items-start gap-3">
						<div className="flex items-center justify-center w-10 h-10 rounded-xl bg-primary/10 text-primary shrink-0">
							<span className="font-bold text-xs">#</span>
						</div>
						<div className="min-w-0 flex-1">
							<p className="text-xs font-medium text-base-content/60 uppercase tracking-wide mb-1">
								ID анализа
							</p>
							<p className="text-sm font-semibold text-base-content break-all">{analysis.id}</p>
						</div>
					</div>
					{analysis.product && (
						<div className="flex items-start gap-3">
							<div className="flex items-center justify-center w-10 h-10 rounded-xl bg-primary/10 text-primary shrink-0">
								<FaBox size={16} />
							</div>
							<div className="min-w-0 flex-1">
								<p className="text-xs font-medium text-base-content/60 uppercase tracking-wide mb-1">
									Продукт
								</p>
								<p className="text-sm font-semibold text-base-content wrap-break-word">
									{translateProductName(analysis.product)}
								</p>
							</div>
						</div>
					)}
					<div className="flex items-start gap-3">
						<div className="flex items-center justify-center w-10 h-10 rounded-xl bg-primary/10 text-primary shrink-0">
							<FaClock size={16} />
						</div>
						<div className="min-w-0 flex-1">
							<p className="text-xs font-medium text-base-content/60 uppercase tracking-wide mb-1">
								Дата и время
							</p>
							<p className="text-sm font-semibold text-base-content wrap-break-word">
								{formatDate(analysis.date_time)}
							</p>
						</div>
					</div>
				</div>
			</AnalysisSheetSection>

			<AnalysisSheetSection
				title="Изображения анализа"
				subtitle="Исходные снимки, отправленные на анализ"
				icon={<IoImageOutline size={20} />}
				accent="info"
				headerExtra={
					imageUrls.length > 0 ? (
						<span className="badge badge-primary badge-sm">
							{imageUrls.length} {imageUrls.length === 1 ? "фото" : "фото"}
						</span>
					) : undefined
				}
			>
				{imageUrls.length > 0 ? (
					<div className="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-3">
						{imageUrls.map((imageUrl, index) => (
							<div
								key={`${analysis.id}-image-${index}`}
								className="relative overflow-hidden rounded-2xl border border-base-200 bg-base-200 aspect-square"
							>
								<ImageFullscreen
									src={imageUrl}
									alt={`Анализ №${analysis.id} - Изображение ${index + 1}`}
									className="w-full h-full object-cover cursor-pointer"
									isClickable={true}
								/>
								<div className="absolute bottom-2 right-2">
									<span className="badge badge-xs bg-base-100/90 backdrop-blur-sm text-base-content border border-base-200">
										{index + 1}
									</span>
								</div>
							</div>
						))}
					</div>
				) : (
					<div className="flex flex-col items-center justify-center py-12 text-center">
						<div className="w-16 h-16 rounded-full bg-base-200 flex items-center justify-center mb-3">
							<IoImageOutline size={32} className="text-base-content/40" />
						</div>
						<p className="text-sm font-medium text-base-content/70">Изображения отсутствуют</p>
						<p className="text-xs text-base-content/50 mt-1">Загрузите изображения для анализа</p>
					</div>
				)}
			</AnalysisSheetSection>

			<AnalysisSheetSection
				title="Обработанные изображения"
				subtitle="Результаты после прохода по конвейеру анализа"
				icon={<IoImageOutline size={20} />}
				accent="success"
				headerExtra={
					processedImages.length > 0 ? (
						<span className="badge badge-success badge-sm">
							{processedImages.length} {processedImages.length === 1 ? "фото" : "фото"}
						</span>
					) : undefined
				}
			>
				{processedImages.length > 0 ? (
					<div className="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-3">
						{processedImages.map((imageUrl, index) => (
							<div
								key={`${analysis.id}-processed-${index}`}
								className="relative overflow-hidden rounded-2xl border border-base-200 bg-base-200 aspect-square"
							>
								<ImageFullscreen
									src={imageUrl}
									alt={`Анализ №${analysis.id} - Обработанное изображение ${index + 1}`}
									className="w-full h-full object-cover cursor-pointer"
									isClickable={true}
								/>
								<div className="absolute bottom-2 right-2">
									<span className="badge badge-xs bg-base-100/90 backdrop-blur-sm text-base-content border border-base-200">
										{index + 1}
									</span>
								</div>
							</div>
						))}
					</div>
				) : (
					<div className="flex flex-col items-center justify-center py-12 text-center">
						<div className="w-16 h-16 rounded-full bg-base-200 flex items-center justify-center mb-3">
							<IoImageOutline size={32} className="text-base-content/40" />
						</div>
						<p className="text-sm font-medium text-base-content/70">
							Обработанные изображения отсутствуют
						</p>
						<p className="text-xs text-base-content/50 mt-1">
							Обработанные изображения появятся после анализа
						</p>
					</div>
				)}
			</AnalysisSheetSection>
		</div>
	)
}

export default MainTab
