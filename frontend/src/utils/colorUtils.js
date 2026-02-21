const clamp = (value, min = 0, max = 1) => Math.min(max, Math.max(min, value))

const hexToRgb = (hexColor) => {
  const hex = hexColor.replace('#', '')
  const value = parseInt(hex.length === 3 ? hex.split('').map((c) => c + c).join('') : hex, 16)
  return {
    r: (value >> 16) & 0xff,
    g: (value >> 8) & 0xff,
    b: value & 0xff
  }
}

const rgbToHex = (r, g, b) => {
  const toHex = (v) => Math.round(v).toString(16).padStart(2, '0').toUpperCase()
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`
}

const srgbToLinear = (channel) => {
  const c = channel / 255
  return c <= 0.04045 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
}

const linearToSrgb = (channel) => {
  const c = clamp(channel)
  return c <= 0.0031308 ? c * 12.92 : 1.055 * Math.pow(c, 1 / 2.4) - 0.055
}

export const getLuminance = (hexColor) => {
  const { r, g, b } = hexToRgb(hexColor)
  const rl = srgbToLinear(r)
  const gl = srgbToLinear(g)
  const bl = srgbToLinear(b)
  return 0.2126 * rl + 0.7152 * gl + 0.0722 * bl
}

export const getContrastRatio = (color1, color2) => {
  const lum1 = getLuminance(color1)
  const lum2 = getLuminance(color2)
  const lighter = Math.max(lum1, lum2)
  const darker = Math.min(lum1, lum2)
  return (lighter + 0.05) / (darker + 0.05)
}

export const getContrastLevel = (ratio) => {
  if (ratio >= 7) return 'AAA'
  if (ratio >= 4.5) return 'AA'
  return 'FAIL'
}

const applyMatrix = (hexColor, matrix) => {
  const { r, g, b } = hexToRgb(hexColor)
  const rl = srgbToLinear(r)
  const gl = srgbToLinear(g)
  const bl = srgbToLinear(b)

  const r2 = matrix[0][0] * rl + matrix[0][1] * gl + matrix[0][2] * bl
  const g2 = matrix[1][0] * rl + matrix[1][1] * gl + matrix[1][2] * bl
  const b2 = matrix[2][0] * rl + matrix[2][1] * gl + matrix[2][2] * bl

  const sr = linearToSrgb(r2) * 255
  const sg = linearToSrgb(g2) * 255
  const sb = linearToSrgb(b2) * 255

  return rgbToHex(sr, sg, sb)
}

const MATRICES = {
  deuteranopia: [
    [0.625, 0.375, 0],
    [0.7, 0.3, 0],
    [0, 0.3, 0.7]
  ],
  protanopia: [
    [0.56667, 0.43333, 0],
    [0.55833, 0.44167, 0],
    [0, 0.24167, 0.75833]
  ],
  tritanopia: [
    [0.95, 0.05, 0],
    [0, 0.43333, 0.56667],
    [0, 0.475, 0.525]
  ],
  achromatopsia: [
    [0.299, 0.587, 0.114],
    [0.299, 0.587, 0.114],
    [0.299, 0.587, 0.114]
  ]
}

export const simulateDeuteranopia = (hexColor) => applyMatrix(hexColor, MATRICES.deuteranopia)

export const simulateProtanopia = (hexColor) => applyMatrix(hexColor, MATRICES.protanopia)

export const simulateTritanopia = (hexColor) => applyMatrix(hexColor, MATRICES.tritanopia)

export const simulateAchromatopsia = (hexColor) => applyMatrix(hexColor, MATRICES.achromatopsia)

/**
 * 将 HEX 颜色转换为 HSL
 * @param {string} hexColor - HEX颜色值 (如 #FF5733)
 * @returns {{ h: number, s: number, l: number }} HSL对象，h为0-360度，s和l为0-100百分比
 */
export const hexToHSL = (hexColor) => {
  const { r, g, b } = hexToRgb(hexColor)
  const rNorm = r / 255
  const gNorm = g / 255
  const bNorm = b / 255

  const max = Math.max(rNorm, gNorm, bNorm)
  const min = Math.min(rNorm, gNorm, bNorm)
  const delta = max - min

  let h = 0
  let s = 0
  const l = (max + min) / 2

  if (delta !== 0) {
    s = l > 0.5 ? delta / (2 - max - min) : delta / (max + min)

    switch (max) {
      case rNorm:
        h = ((gNorm - bNorm) / delta + (gNorm < bNorm ? 6 : 0)) / 6
        break
      case gNorm:
        h = ((bNorm - rNorm) / delta + 2) / 6
        break
      case bNorm:
        h = ((rNorm - gNorm) / delta + 4) / 6
        break
    }
  }

  return {
    h: Math.round(h * 360),
    s: Math.round(s * 100),
    l: Math.round(l * 100)
  }
}

/**
 * 计算两个颜色的 HSL 差值
 * @param {string} color1 - 第一个颜色（HEX）
 * @param {string} color2 - 第二个颜色（HEX）
 * @returns {{ dH: number, dS: number, dL: number }} HSL差值对象
 */
export const getHSLDifference = (color1, color2) => {
  const hsl1 = hexToHSL(color1)
  const hsl2 = hexToHSL(color2)

  // 色相环形差值计算（取较短路径）
  let dH = hsl2.h - hsl1.h
  if (dH > 180) dH -= 360
  if (dH < -180) dH += 360

  return {
    dH: Math.round(dH),
    dS: hsl2.s - hsl1.s,
    dL: hsl2.l - hsl1.l
  }
}
