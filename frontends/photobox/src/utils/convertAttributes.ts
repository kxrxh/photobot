export function convertAttribute(attribute: string): string {
	const directMapping: Record<string, string> = {
		l: "L",
		w: "W",
		sq: "Sq",
		s: "S",
		v: "V",
		h: "H",
		r: "R",
		g: "G",
		b: "B",
		pr: "Pr",
		brt: "Brt",
		solid: "Solid",
		l_w: "L/W",
		l_avg: "L/Avg",
		w_avg: "W/Avg",
		min_h: "Min_H",
		min_s: "Min_S",
		min_v: "Min_V",
		max_h: "Max_H",
		max_s: "Max_S",
		max_v: "Max_V",
		min_r: "Min_R",
		min_g: "Min_G",
		min_b: "Min_B",
		max_r: "Max_R",
		max_g: "Max_G",
		max_b: "Max_B",
		r_avg: "R/Avg",
		g_avg: "G/Avg",
		b_avg: "B/Avg",
		h_avg: "H/Avg",
		s_avg: "S/Avg",
		v_avg: "V/Avg",
		brt_avg: "Brt/Avg",
		sq_sqcrl: "Sq/SqCrl",
		median_r: "median_R",
		median_g: "median_G",
		median_b: "median_B",
		median_h: "median_H",
		median_s: "median_S",
		median_v: "median_V",
		l_m: "L/M",
		w_m: "W/M",
		r_m: "R/M",
		g_m: "G/M",
		b_m: "B/M",
		h_m: "H/M",
		s_m: "S/M",
		v_m: "V/M",
		brt_m: "Brt/M",
		m_r: "R/M",
		m_g: "G/M",
		m_b: "B/M",
		m_h: "H/M",
		m_s: "S/M",
		m_v: "V/M",
	}

	const lower = attribute.toLowerCase()

	if (directMapping[lower]) {
		return directMapping[lower]
	}

	if (attribute.length === 1 && /[a-z]/.test(attribute)) {
		return attribute.toUpperCase()
	}

	return attribute
}
