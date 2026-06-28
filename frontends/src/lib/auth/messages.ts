export const REFRESH_EXPIRED_MESSAGE = "Сессия истекла. Войдите снова."

export const MESSENGER_RELOAD_MESSAGE = "Сессия истекла. Закройте и откройте приложение заново."

export const GENERIC_AUTH_ERROR_MESSAGE = "Не удалось выполнить вход. Попробуйте снова."

export const NETWORK_ERROR_MESSAGE =
	"Сервер не отвечает. Пожалуйста, проверьте ваше интернет-соединение или попробуйте позже."

export const REQUEST_TIMEOUT_MESSAGE =
	"Превышено время ожидания. Проверьте интернет-соединение; при загрузке файлов попробуйте меньше файлов за раз или более стабильную сеть."

export const AUTH_UNAUTHORIZED_MESSAGE = "Ошибка авторизации."

export const FORBIDDEN_MESSAGE = "Доступ запрещён."

export const NOT_FOUND_MESSAGE = "Не найдено."

export const GENERIC_REQUEST_ERROR_MESSAGE = "Произошла ошибка при запросе."

export const SERVER_ERROR_MESSAGE = "Временная ошибка сервера. Попробуйте позже."

export const TELEGRAM_VALIDATION_ERROR_MESSAGE =
	"Ошибка проверки данных Telegram. Перезапустите приложение."

export const INIT_DATA_UNAVAILABLE_MESSAGE =
	"Данные мессенджера недоступны. Перезапустите приложение."

export const UNKNOWN_ERROR_MESSAGE = "Произошла непредвиденная ошибка."

export const AUTH_API_MESSAGES: Record<string, string> = {
	"invalid credentials": "Неверный логин или пароль",
	"missing login credentials": "Введите логин и пароль",
	"missing password credentials": "Введите логин и пароль",
	"invalid refresh token": REFRESH_EXPIRED_MESSAGE,
	"invalid token type": REFRESH_EXPIRED_MESSAGE,
	"refresh token expired or already used": REFRESH_EXPIRED_MESSAGE,
	"telegram data validation failed":
		"Не удалось подтвердить Telegram-сессию. Перезапустите приложение.",
	"max data validation failed": "Не удалось подтвердить MAX-сессию. Перезапустите приложение.",
	"unsupported messenger platform": "Неподдерживаемая платформа мессенджера",
	"unsupported grant type": "Неподдерживаемый способ входа",
	"user not found": "Пользователь не найден",
	"invalid login format": "Некорректный формат логина",
	"invalid reset request": "Некорректный запрос на сброс пароля",
	"reset code not found or expired": "Код сброса не найден или истёк",
	"invalid reset code": "Некорректный код сброса",
	"invalid recovery code": "Некорректный код восстановления",
	"cannot link account to itself":
		"Этот код выдан для этого же аккаунта. Получите код в другом приложении.",
	"link code not found or expired": "Код не найден или истёк. Запросите новый код.",
	"code user not found": "Пользователь по коду не найден",
	"init data does not match the authenticated user":
		"Ошибка подтверждения аккаунта. Перезапустите приложение.",
	"authenticated user not found": "Пользователь не найден. Перезапустите приложение.",
	"link code must be 6 digits": "Код должен состоять из 6 цифр",
	"login already taken": "Логин уже занят",
	"логин уже занят": "Логин уже занят",
	"web login already set": "Веб-доступ для этого аккаунта уже настроен",
	"password must be at least 6 characters": "Пароль должен содержать не менее 6 символов",
	"login must be 3-32 characters: lowercase letters, digits, underscore, hyphen":
		"Логин: 3–32 символа, только a-z, 0-9, _ и -",
	"номер телефона уже зарегистрирован": "Номер телефона уже зарегистрирован",
	"данные уже используются. проверьте введённые значения.":
		"Данные уже используются. Проверьте введённые значения.",
}

export function isRefreshTokenError(message: string): boolean {
	const normalized = message.trim().toLowerCase()
	return (
		normalized.includes("refresh token") ||
		normalized === "invalid token type" ||
		normalized.includes("token expired or already used")
	)
}

export function toUserFacingAuthMessage(message: string): string {
	const trimmed = message.trim()
	const normalized = trimmed.toLowerCase()
	if (isRefreshTokenError(message)) return REFRESH_EXPIRED_MESSAGE
	if (AUTH_API_MESSAGES[normalized]) return AUTH_API_MESSAGES[normalized]
	if (/[а-яё]/i.test(trimmed)) return trimmed
	return GENERIC_AUTH_ERROR_MESSAGE
}
