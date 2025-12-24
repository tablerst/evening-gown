import type { GlobalThemeOverrides } from 'naive-ui'

// Admin-only Naive UI theme overrides.
// Keep it aligned with STYLE.md (sharp edges, minimal shadows, monochrome + deep navy accent).
export const adminThemeOverrides: GlobalThemeOverrides = {
    common: {
        primaryColor: '#000226',
        primaryColorHover: '#000226',
        primaryColorPressed: '#000226',
        primaryColorSuppl: '#000226',

        borderColor: '#E2E8F0',
        dividerColor: '#E2E8F0',

        baseColor: '#FFFFFF',
        bodyColor: '#FFFFFF',
        cardColor: '#FFFFFF',
        modalColor: '#FFFFFF',
        popoverColor: '#FFFFFF',

        textColorBase: '#000000',
        textColor1: 'rgba(0,0,0,0.90)',
        textColor2: 'rgba(0,0,0,0.70)',
        textColor3: 'rgba(0,0,0,0.55)',

        borderRadius: '0px',

        fontFamily: 'Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, "Apple Color Emoji", "Segoe UI Emoji"',
        fontFamilyMono: '"JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
    },

    Layout: {
        siderColor: '#FFFFFF',
        headerColor: '#FFFFFF',
        color: '#FFFFFF',
    },

    Card: {
        borderRadius: '0px',
        boxShadow: 'none',
    },

    Button: {
        borderRadiusMedium: '0px',
        borderRadiusSmall: '0px',
        borderRadiusTiny: '0px',
        textColor: 'rgba(0,0,0,0.90)',
        textColorHover: 'rgba(0,0,0,1)',
        textColorPressed: 'rgba(0,0,0,1)',
    },

    Input: {
        borderRadius: '0px',
    },

    DataTable: {
        thColor: '#FFFFFF',
        tdColor: '#FFFFFF',
        thTextColor: 'rgba(0,0,0,0.70)',
    },

    Menu: {
        itemBorderRadius: '0px',
    },
}
