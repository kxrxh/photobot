import { Link } from "@tanstack/react-router"
import { useCallback, useEffect, useId, useState } from "react"
import { FaPaste } from "react-icons/fa"
import {
	detectMessengerPlatform,
	getMe,
	linkWithCode,
	linkWithCodeFromWeb,
	requestLinkCode,
} from "@/api/auth"
import type { AuthServiceUser } from "@/api/auth/types"
import { ApiError } from "@/api/types"
import { Button } from "@/components/common/ui/Button"
import { Input } from "@/components/common/ui/Input"
import { useAuth } from "@/contexts/AuthContext"
import { useAlert } from "@/hooks/useAlert"
import { useMessenger } from "@/hooks/useMessenger"
import { getLinkErrorMessage } from "@/lib/auth/linkErrors"
import { isWebAuthMode } from "@/lib/auth/mode"

type LinkSnapshot = Pick<AuthServiceUser, "login" | "telegram_id" | "max_id">

function LinkStatusRow({
	label,
	linked,
	detail,
}: {
	label: string
	linked: boolean
	detail?: string
}) {
	return (
		<div className="flex items-start justify-between gap-3 py-2.5 border-b border-base-300/50 last:border-0">
			<div className="min-w-0">
				<p className="text-sm font-medium">{label}</p>
				{detail ? <p className="text-xs text-base-content/55 mt-0.5">{detail}</p> : null}
			</div>
			<span className={`badge badge-sm shrink-0 ${linked ? "badge-success" : "badge-ghost"}`}>
				{linked ? "Да" : "Нет"}
			</span>
		</div>
	)
}

export function AccountLinkingSection() {
	const collapseId = useId()
	const { userId, applyTokenResponse, logout } = useAuth()
	const { showSuccess, showError: showErrorAlert } = useAlert()
	const { initData, hapticFeedback } = useMessenger()

	const [links, setLinks] = useState<LinkSnapshot | null>(null)
	const [linksLoaded, setLinksLoaded] = useState(false)
	const [linkCode, setLinkCode] = useState("")
	const [linkCodeDisplay, setLinkCodeDisplay] = useState<string | null>(null)
	const [isLoading, setIsLoading] = useState(false)

	const refreshLinks = useCallback(async () => {
		try {
			const data = await getMe()
			setLinks({
				login: data.login,
				telegram_id: data.telegram_id,
				max_id: data.max_id,
			})
		} catch (err) {
			if (ApiError.isApiError(err) && err.isNotFound()) {
				logout()
				showErrorAlert(
					"Аккаунт объединён с другим профилем. Войдите снова через Telegram или с логином объединённого аккаунта."
				)
			}
			throw err
		}
	}, [logout, showErrorAlert])

	useEffect(() => {
		if (!userId) return
		refreshLinks()
			.catch(() => setLinks({ login: null, telegram_id: null, max_id: null }))
			.finally(() => setLinksLoaded(true))
	}, [userId, refreshLinks])

	const platform = initData ? detectMessengerPlatform(initData) : "telegram"
	const inMessenger = Boolean(initData)
	const webMode = isWebAuthMode()

	const hasWeb = Boolean(links?.login?.trim())
	const hasTelegram = links?.telegram_id != null
	const hasMax = links?.max_id != null
	const needsMessenger = !hasTelegram || !hasMax
	const needsMessengerCrossLink = (hasTelegram && !hasMax) || (!hasTelegram && hasMax)
	const showWebLinkCode = webMode && !inMessenger && needsMessenger
	const showWebCodeInput = webMode && !inMessenger && needsMessenger
	const showMessengerCodeInput = inMessenger && (needsMessenger || needsMessengerCrossLink)
	const showWebAccessSetup = inMessenger && !hasWeb
	const linkedCount = [hasWeb, hasTelegram, hasMax].filter(Boolean).length
	const needsAction = linkedCount < 3

	const handleRequestLinkCode = async () => {
		setIsLoading(true)
		setLinkCodeDisplay(null)
		try {
			const result = await requestLinkCode()
			setLinkCodeDisplay(result.code)
			try {
				await navigator.clipboard.writeText(result.code)
				showSuccess("Код скопирован")
			} catch {
				showSuccess("Код получен — действует 5 минут")
			}
			hapticFeedback?.notificationOccurred?.("success")
		} catch (err) {
			showErrorAlert(getLinkErrorMessage(err, "Не удалось получить код"))
		} finally {
			setIsLoading(false)
		}
	}

	const handlePasteCode = async () => {
		try {
			const text = await navigator.clipboard.readText()
			const digits = text.replace(/\D/g, "").slice(0, 6)
			if (digits) {
				setLinkCode(digits)
				hapticFeedback?.notificationOccurred?.("success")
			}
		} catch {
			showErrorAlert("Не удалось вставить из буфера")
		}
	}

	const handleWebLinkWithCode = async (e: React.FormEvent) => {
		e.preventDefault()
		if (linkCode.length !== 6) {
			showErrorAlert("Введите 6-значный код")
			return
		}
		setIsLoading(true)
		try {
			const result = await linkWithCodeFromWeb(linkCode.trim())
			if (result.access_token && result.refresh_token) {
				await applyTokenResponse({
					access_token: result.access_token,
					refresh_token: result.refresh_token,
					roles: result.roles ?? [],
				})
			}
			await refreshLinks()
			showSuccess(result.message || "Аккаунт привязан")
			setLinkCode("")
			setLinkCodeDisplay(null)
		} catch (err) {
			showErrorAlert(getLinkErrorMessage(err))
		} finally {
			setIsLoading(false)
		}
	}

	const handleMessengerLinkWithCode = async (e: React.FormEvent) => {
		e.preventDefault()
		if (!initData || linkCode.length !== 6) {
			showErrorAlert("Введите 6-значный код")
			return
		}
		setIsLoading(true)
		try {
			const result = await linkWithCode(linkCode.trim(), initData, platform)
			if (result.access_token && result.refresh_token) {
				await applyTokenResponse({
					access_token: result.access_token,
					refresh_token: result.refresh_token,
					roles: result.roles ?? [],
				})
			}
			await refreshLinks()
			showSuccess(result.message || "Аккаунт привязан")
			setLinkCode("")
			setLinkCodeDisplay(null)
			hapticFeedback?.notificationOccurred?.("success")
		} catch (err) {
			showErrorAlert(getLinkErrorMessage(err))
		} finally {
			setIsLoading(false)
		}
	}

	return (
		<div className="collapse collapse-arrow rounded-box border border-base-300 bg-base-200/30">
			<input id={collapseId} type="checkbox" defaultChecked={needsAction || !linksLoaded} />
			<label htmlFor={collapseId} className="collapse-title min-h-0 py-3 pr-8">
				<div className="flex flex-col items-start gap-0.5">
					<span className="text-sm font-medium">Связанные аккаунты</span>
					<span className="text-xs font-normal text-base-content/55">
						{linksLoaded
							? linkedCount === 3
								? "Всё привязано"
								: `${linkedCount} из 3`
							: "Загрузка…"}
					</span>
				</div>
			</label>
			<div className="collapse-content">
				<div className="flex flex-col gap-4 px-1 pb-4">
					{!linksLoaded ? (
						<div className="flex justify-center py-4">
							<span className="loading loading-spinner loading-md" />
						</div>
					) : (
						<>
							<div className="rounded-box bg-base-100/80 px-3">
								<LinkStatusRow
									label="Браузер (логин и пароль)"
									linked={hasWeb}
									detail={hasWeb ? (links?.login ?? undefined) : undefined}
								/>
								<LinkStatusRow label="Telegram" linked={hasTelegram} />
								<LinkStatusRow label="MAX" linked={hasMax} />
							</div>

							{showWebAccessSetup ? (
								<div className="rounded-box bg-base-100/60 p-3 flex flex-col gap-2">
									<p className="text-sm font-medium">Вход через браузер</p>
									<p className="text-xs text-base-content/65">
										Создайте логин и пароль для этого же аккаунта — без кодов и переноса данных.
									</p>
									<Link to="/settings/web-access" className="btn btn-primary btn-sm">
										Создать логин и пароль
									</Link>
								</div>
							) : null}

							{showWebCodeInput ? (
								<div className="rounded-box bg-base-100/60 p-3 flex flex-col gap-3">
									<p className="text-sm font-medium">Ввести код из мессенджера</p>
									<p className="text-xs text-base-content/65">
										Код, сгенерированный в Telegram или MAX (Профиль → Связанные аккаунты).
									</p>
									<form onSubmit={handleWebLinkWithCode} className="flex flex-col gap-2">
										<div className="flex gap-2">
											<Input
												type="text"
												inputMode="numeric"
												maxLength={6}
												placeholder="000000"
												size="sm"
												className="flex-1 text-center font-mono tracking-widest"
												value={linkCode}
												onChange={(e) => setLinkCode(e.target.value.replace(/\D/g, ""))}
												disabled={isLoading}
												autoComplete="one-time-code"
											/>
											<Button
												type="button"
												variant="outline"
												size="sm"
												className="btn-square shrink-0"
												onClick={handlePasteCode}
												disabled={isLoading}
												title="Вставить"
											>
												<FaPaste />
											</Button>
										</div>
										<Button
											type="submit"
											variant="primary"
											size="sm"
											fullWidth
											loading={isLoading}
											disabled={linkCode.length !== 6}
										>
											Привязать
										</Button>
									</form>
								</div>
							) : null}

							{showWebLinkCode ? (
								<div className="rounded-box bg-base-100/60 p-3 flex flex-col gap-3">
									<p className="text-sm font-medium">Привязать Telegram или MAX</p>
									<ol className="list-decimal list-inside text-xs text-base-content/70 space-y-1">
										<li>Нажмите «Сгенерировать код»</li>
										<li>Откройте бота в Telegram или MAX</li>
										<li>Профиль → «Связанные аккаунты» → введите код</li>
									</ol>
									{!linkCodeDisplay ? (
										<Button
											type="button"
											variant="primary"
											size="sm"
											fullWidth
											loading={isLoading}
											onClick={handleRequestLinkCode}
										>
											Сгенерировать код
										</Button>
									) : (
										<div className="flex flex-col gap-2">
											<button
												type="button"
												className="w-full bg-base-300/60 rounded-box p-3 text-center"
												onClick={() => {
													navigator.clipboard.writeText(linkCodeDisplay)
													showSuccess("Код скопирован")
												}}
											>
												<p className="font-mono text-2xl font-semibold tracking-[0.3em]">
													{linkCodeDisplay}
												</p>
												<p className="text-xs text-base-content/60 mt-1">
													Нажмите, чтобы скопировать · 5 мин
												</p>
											</button>
											<Button
												type="button"
												variant="ghost"
												size="sm"
												onClick={handleRequestLinkCode}
												disabled={isLoading}
											>
												Новый код
											</Button>
										</div>
									)}
								</div>
							) : null}

							{showMessengerCodeInput ? (
								<div className="rounded-box bg-base-100/60 p-3 flex flex-col gap-3">
									<p className="text-sm font-medium">Ввести код</p>
									<p className="text-xs text-base-content/65">
										{needsMessengerCrossLink
											? `Код из ${platform === "telegram" ? "MAX" : "Telegram"} или с сайта.`
											: "Код, сгенерированный в браузере (Профиль → Связанные аккаунты)."}
									</p>
									<form onSubmit={handleMessengerLinkWithCode} className="flex flex-col gap-2">
										<div className="flex gap-2">
											<Input
												type="text"
												inputMode="numeric"
												maxLength={6}
												placeholder="000000"
												size="sm"
												className="flex-1 text-center font-mono tracking-widest"
												value={linkCode}
												onChange={(e) => setLinkCode(e.target.value.replace(/\D/g, ""))}
												disabled={isLoading}
												autoComplete="one-time-code"
											/>
											<Button
												type="button"
												variant="outline"
												size="sm"
												className="btn-square shrink-0"
												onClick={handlePasteCode}
												disabled={isLoading}
												title="Вставить"
											>
												<FaPaste />
											</Button>
										</div>
										<Button
											type="submit"
											variant="primary"
											size="sm"
											fullWidth
											loading={isLoading}
											disabled={linkCode.length !== 6}
										>
											Привязать
										</Button>
									</form>
									{needsMessengerCrossLink ? (
										<>
											<div className="divider my-0 text-xs">или отправить код отсюда</div>
											<p className="text-xs text-base-content/65">
												Сгенерируйте код и введите его в{" "}
												{platform === "telegram" ? "MAX" : "Telegram"} (Профиль → Связанные
												аккаунты).
											</p>
											{!linkCodeDisplay ? (
												<Button
													type="button"
													variant="outline"
													size="sm"
													fullWidth
													loading={isLoading}
													onClick={handleRequestLinkCode}
												>
													Сгенерировать код
												</Button>
											) : (
												<button
													type="button"
													className="w-full bg-base-300/60 rounded-box p-3 text-center"
													onClick={() => {
														navigator.clipboard.writeText(linkCodeDisplay)
														showSuccess("Код скопирован")
													}}
												>
													<p className="font-mono text-xl font-semibold tracking-[0.25em]">
														{linkCodeDisplay}
													</p>
												</button>
											)}
										</>
									) : null}
								</div>
							) : null}

							{webMode && !inMessenger && !hasWeb ? (
								<p className="text-xs text-base-content/60 text-center">
									Новый веб-аккаунт —{" "}
									<Link to="/register" className="link link-primary">
										регистрация
									</Link>
									. Уже есть мессенджер-аккаунт — создайте веб-доступ в мини-приложении.
								</p>
							) : null}
						</>
					)}
				</div>
			</div>
		</div>
	)
}
