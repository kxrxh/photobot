import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useCallback, useEffect, useId, useRef, useState } from "react"
import { getMe, updateMe } from "@/api/auth"
import { Button } from "@/components/common/ui/Button"
import { AccountLinkingSection } from "@/components/profile/AccountLinkingSection"
import { useAuth } from "@/contexts/AuthContext"
import { useAlert } from "@/hooks/useAlert"
import { isWebAuthMode } from "@/lib/auth/mode"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"

type FieldName = "name" | "organization" | "inn" | "phone"
type ValidationState = {
	isNameValid: boolean
	isOrganizationValid: boolean
	isInnValid: boolean
	isPhoneValid: boolean
}
type TouchedState = {
	[K in FieldName]: boolean
}

export const Route = createFileRoute("/_authenticated/profile")({
	component: RouteComponent,
})

function RouteComponent() {
	const { userId, state, logout } = useAuth()
	const { showSuccess, showWarning, showError: showErrorAlert } = useAlert()
	const navigate = useNavigate()
	const [isLoading, setIsLoading] = useState(false)
	const [isDataLoading, setIsDataLoading] = useState(true)
	const dataLoadedRef = useRef(false)
	const initialFormDataRef = useRef<{
		name: string
		organization: string
		inn: string
		phone: string
	} | null>(null)

	const nameId = useId()
	const organizationId = useId()
	const innId = useId()
	const phoneId = useId()

	const formatPhoneNumber = useCallback((value: string) => {
		const phoneNumber = value.replace(/\D/g, "")
		if (phoneNumber.length === 0) return ""
		if (phoneNumber.length <= 1) return "+7"
		if (phoneNumber.length <= 4) return `+7 (${phoneNumber.slice(1)}`
		if (phoneNumber.length <= 7) return `+7 (${phoneNumber.slice(1, 4)}) ${phoneNumber.slice(4)}`
		if (phoneNumber.length <= 9)
			return `+7 (${phoneNumber.slice(1, 4)}) ${phoneNumber.slice(4, 7)}-${phoneNumber.slice(7)}`
		return `+7 (${phoneNumber.slice(1, 4)}) ${phoneNumber.slice(4, 7)}-${phoneNumber.slice(7, 9)}-${phoneNumber.slice(9, 11)}`
	}, [])

	const validateInn = useCallback((inn: string) => {
		return /^(\d{10}|\d{12})$/.test(inn)
	}, [])

	const validatePhone = useCallback((phone: string) => {
		return phone.length === 18
	}, [])

	const [formData, setFormData] = useState({
		name: "",
		organization: "",
		inn: "",
		phone: "",
	})

	const [validation, setValidation] = useState<ValidationState>({
		isNameValid: true,
		isOrganizationValid: true,
		isInnValid: true,
		isPhoneValid: true,
	})

	const [touched, setTouched] = useState<TouchedState>({
		name: false,
		organization: false,
		inn: false,
		phone: false,
	})

	useEffect(() => {
		const loadUserData = async () => {
			if (state === "loading") return
			if (state !== "authenticated" || !userId) {
				setIsDataLoading(false)
				return
			}
			if (dataLoadedRef.current) {
				setIsDataLoading(false)
				return
			}
			try {
				const userData = await getMe()
				const formattedPhone = formatPhoneNumber(userData.phone_number || "")
				const loaded = {
					name: userData.full_name || "",
					organization: userData.organization_name || "",
					inn: userData.inn || "",
					phone: formattedPhone,
				}
				setFormData(loaded)
				initialFormDataRef.current = loaded
				setValidation({
					isNameValid: true,
					isOrganizationValid: true,
					isInnValid: !userData.inn || validateInn(userData.inn),
					isPhoneValid: !formattedPhone || validatePhone(formattedPhone),
				})
				dataLoadedRef.current = true
			} catch (error) {
				log.devError("Profile: Error loading user data:", error)
				showErrorAlert(getUserFacingErrorMessage(error))
			} finally {
				setIsDataLoading(false)
			}
		}
		loadUserData()
	}, [state, userId, showErrorAlert, formatPhoneNumber, validateInn, validatePhone])

	const getValidationKey = (field: FieldName): keyof ValidationState =>
		`is${field.charAt(0).toUpperCase() + field.slice(1)}Valid` as keyof ValidationState

	const handleInputChange = (field: FieldName) => (e: React.ChangeEvent<HTMLInputElement>) => {
		const value = e.target.value
		const newValue = field === "phone" ? formatPhoneNumber(value) : value
		setFormData((prev) => ({ ...prev, [field]: newValue }))
		setValidation((prev) => ({
			...prev,
			[getValidationKey(field)]:
				field === "phone"
					? !newValue || validatePhone(newValue)
					: field === "inn"
						? !newValue || validateInn(newValue)
						: true,
		}))
		setTouched((prev) => ({ ...prev, [field]: true }))
	}

	const getInputClassName = (field: FieldName) => {
		const isValid = validation[getValidationKey(field)]
		const isTouched = touched[field]
		const showFieldError = field === "inn" || field === "phone" ? isTouched && !isValid : false
		return `w-full input input-bordered ${showFieldError ? "input-error" : ""}`
	}

	const showFieldError = (field: FieldName) =>
		(field === "inn" || field === "phone") && touched[field] && !validation[getValidationKey(field)]

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault()
		if (!userId) {
			showErrorAlert("Ошибка: пользователь не авторизован")
			return
		}
		if (!Object.values(validation).every(Boolean)) {
			showWarning("Пожалуйста, исправьте ошибки в форме (ИНН и телефон при заполнении)")
			return
		}
		const initial = initialFormDataRef.current
		const unchanged =
			initial &&
			formData.name.trim() === initial.name.trim() &&
			formData.organization.trim() === initial.organization.trim() &&
			formData.inn.trim() === initial.inn.trim() &&
			formData.phone === initial.phone
		if (unchanged) {
			navigate({ to: "/" })
			return
		}
		setIsLoading(true)
		try {
			await updateMe({
				full_name: formData.name.trim() || undefined,
				organization_name: formData.organization.trim() || undefined,
				inn: formData.inn.trim() || undefined,
				phone_number:
					formData.phone.replace(/\D/g, "").length >= 11
						? formData.phone.replace(/\D/g, "")
						: undefined,
			})
			showSuccess("Профиль успешно обновлён")
			navigate({ to: "/" })
		} catch (error) {
			log.devError("Profile: Error updating user data:", error)
			showErrorAlert(getUserFacingErrorMessage(error))
		} finally {
			setIsLoading(false)
		}
	}

	if (state === "loading" || isDataLoading) {
		return (
			<div className="flex flex-col w-full max-w-md gap-6 p-4 mx-auto">
				<h2 className="text-2xl font-semibold text-center">Профиль</h2>
				<div className="flex items-center justify-center py-8">
					<span className="loading loading-spinner loading-lg" />
				</div>
			</div>
		)
	}

	if (state !== "authenticated" || !userId) {
		return (
			<div className="flex flex-col w-full max-w-md gap-6 p-4 mx-auto">
				<h2 className="text-2xl font-semibold text-center">Профиль</h2>
				<div className="flex items-center justify-center py-8">
					<div className="text-center">
						<p className="text-error mb-4">Ошибка аутентификации</p>
						<p className="text-sm text-base-content/60">Пожалуйста, войдите в систему заново</p>
					</div>
				</div>
			</div>
		)
	}

	return (
		<div className="flex flex-col w-full max-w-md gap-6 p-4 mx-auto">
			<h2 className="text-2xl font-semibold text-center">Профиль</h2>
			<AccountLinkingSection />
			<form className="flex flex-col gap-2" onSubmit={handleSubmit}>
				<div className="w-full form-control">
					<label htmlFor={nameId} className="label">
						<span className="font-medium label-text">ФИО</span>
					</label>
					<input
						type="text"
						id={nameId}
						className={getInputClassName("name")}
						placeholder="Опционально"
						value={formData.name}
						onChange={handleInputChange("name")}
					/>
				</div>
				<div className="w-full form-control">
					<label htmlFor={organizationId} className="label">
						<span className="font-medium label-text">Организация</span>
					</label>
					<input
						type="text"
						id={organizationId}
						className={getInputClassName("organization")}
						placeholder="Опционально"
						value={formData.organization}
						onChange={handleInputChange("organization")}
					/>
				</div>
				<div className="w-full form-control">
					<label htmlFor={innId} className="label">
						<span className="font-medium label-text">ИНН</span>
					</label>
					<input
						type="text"
						id={innId}
						className={getInputClassName("inn")}
						placeholder="10 или 12 цифр (опционально)"
						maxLength={12}
						value={formData.inn}
						onChange={handleInputChange("inn")}
					/>
					{showFieldError("inn") && (
						<div className="mt-1 text-sm text-error">
							Введите ИНН (10 цифр для организаций, 12 цифр для ИП)
						</div>
					)}
				</div>
				<div className="w-full form-control">
					<label htmlFor={phoneId} className="label">
						<span className="font-medium label-text">Номер телефона</span>
					</label>
					<input
						type="tel"
						id={phoneId}
						className={getInputClassName("phone")}
						placeholder="+7 (999) 999-99-99 (опционально)"
						maxLength={18}
						value={formData.phone}
						onChange={handleInputChange("phone")}
					/>
					{showFieldError("phone") && (
						<div className="mt-1 text-sm text-error">
							Введите корректный номер телефона в формате +7 (XXX) XXX-XX-XX
						</div>
					)}
				</div>
				<button
					className={`w-full mt-2 btn btn-primary ${isLoading ? "loading" : ""}`}
					type="submit"
					disabled={isLoading}
				>
					{isLoading ? "Сохранение..." : "Сохранить"}
				</button>
			</form>
			{isWebAuthMode() ? (
				<Button
					type="button"
					variant="outline"
					fullWidth
					onClick={() => {
						logout()
						navigate({ to: "/login" })
					}}
				>
					Выйти
				</Button>
			) : null}
		</div>
	)
}
