export type SupportedLocale = 'zh' | 'en'

export const messages = {
    en: {
        nav: {
            brandLine1: 'NOIR & ÉCLAT',
            brandLine2: 'FLEURLIS',
            links: {
                process: 'Process',
                brief: 'Brief',
                seasonal: 'Seasonal',
                catalog: 'Catalog',
                appointment: 'Appointment',
            },
            cta: 'Book Private Salon',
            language: {
                toggle: 'Language',
                zh: '中文',
                en: 'EN',
            },
        },
        footer: {
            tagline: 'BY APPOINTMENT ONLY',
            brand: 'FLEURLIS',
            instagram: 'Instagram',
            wechat: 'WeChat',
            email: 'Email',
            legal: '© 2025 FLEURLIS · SUZHOU · EST. 2017',
            icp: '苏ICP备2025229492号-1',
        },
        hero: {
            label: '// INTRODUCTION_THEATRE',
            body: `EST. 2008 · SHUNHE FACTORY
20 YEARS IN CUSTOM & WHOLESALE
WE KNOW QUALITY · MARKET · DELIVERY

CRAFTSMANSHIP IN EVERY STEP
--------------------------------
SOURCE FACTORY · GLOBAL SUPPLY`,
        },
        ticker: {
            release: 'NEW COLLECTION RELEASE · 2025 S/S · WHOLESALE ONLY ·',
            fashionWeek: 'SHANGHAI FASHION WEEK · BOOTH A-12 ·',
            leadTime: 'PRODUCTION LEAD TIME: 14 DAYS ·',
        },
        seasonal: {
            title: 'Seasonal Collection',
            season: '2025 S/S',
            series: 'SERIES {index}',
            lookbook: 'VIEW LOOKBOOK ->',
        },
        product: {
            title: 'New Arrivals',
            total: 'TOTAL: {count}',
            filter: 'FILTER: ALL',
            sort: 'SORT: DATE',
            style: 'STYLE #{id}',
            login: 'LOGIN FOR PRICE',
            add: 'Add to list',
        },
        info: {
            designerTitle: "Designer's Note",
            designerBody:
                'We strip away the unnecessary. No ruffles, no excess. Just structure and silhouette. Our designs are architectural blueprints for the body.',
            designerSign: '- Chief Director',
            appointmentTitle: 'Appointment',
            appointmentBody:
                'Visit our showroom to feel the fabric quality. We offer private viewing sessions for wholesale partners.',
            appointmentCta: 'Book Visit',
            brandTitle: 'The Brand',
            brandStats: {
                est: 'EST.',
                estValue: '2017',
                hq: 'HQ',
                hqValue: 'SUZHOU',
                clients: 'CLIENTS',
                clientsValue: 'GLOBAL B2B',
                capacity: 'CAPACITY',
                capacityValue: '50,000/YR',
            },
        },
        preview: {
            ref: 'REF: 2025-ARCHIVE',
            security: 'SECURE GATEWAY',
            heading: 'Design-Production-Sales',
            subheading: 'Est. 2017 · Full Process Supported',
            enter: '[ Enter Archive ]',
            ready: 'System Ready',
        },
        atelier: {
            eyebrow: 'THE ATELIER',
            title: 'Atelier Flow · Linear Precision',
            lede: 'Present the couture workflow like an interface so clients read fabric weight, sewing steps, and schedules as if browsing product specs.',
            routineTitle: 'Routine Timeline',
            timeline: [
                {
                    title: 'DAY 01 · PATTERN SKETCH',
                    desc: '72h sketching and draping; transparent taffeta calibrates the pattern.',
                },
                {
                    title: 'DAY 04 · MATERIAL CURATION',
                    desc: 'Champagne taffeta × matte pearl organza; log humidity and texture.',
                },
                {
                    title: 'DAY 09 · FITTING',
                    desc: 'Three fittings + laser embroidery alignment; keep a data log for every adjustment.',
                },
            ],
            meta: ['Light transmission test', 'Fragrance pairing', 'Tactile archive'],
            materialLabTitle: 'Material Lab',
            materialLabCopy:
                'Record each material with high-key photography and sync to WebGL silk shader parameters for digital twins, keeping on/offline color temperature aligned.',
            pills: {
                reflectance: '0.78 · pearl',
                roughness: '0.32 · satin',
                colorDrift: 'Cool white ↔ Warm gold',
                windProfile: 'Perlin 0.6',
            },
        },
        couture: {
            eyebrow: 'COUTURE & CUSTOM',
            title: 'Book a FLEURLIS fitting',
            copy:
                'Each fitting happens in a frosted-glass quiet room with dawn-like ambient light. Stylists deliver a look sheet tuned to skin tone and occasion, synced with fragrance notes and fabric texture archives.',
            bullets: [
                'Remote styling session · 45 minutes',
                'Fabric tactile archive + scent kit delivery',
                'Sketch in 72 hours, first fitting in 14 days',
            ],
            badge: 'By Appointment Only',
            formTitle: 'Schedule your fitting',
            formCopy: 'Tell us your city and preferred date; the atelier will arrange the nearest FLEURLIS salon.',
            form: {
                city: 'City',
                date: 'Date',
                placeholder: 'Shanghai / Paris',
                submit: 'Submit request',
            },
        },
        gallery: {
            eyebrow: 'THE GALLERY',
            title: 'Vernissage · White Halo Showcase',
            lede: 'The asymmetric bento grid splits gowns, couture accessories, and material close-ups. Each card floats on the “atelier haze” with a 1px linear frame, revealing high-key highlights.',
            cards: {
                halo: {
                    caption: 'LOOK 01 · STRUCTURE',
                    title: 'Architected Halo',
                    desc: '3D corsetry · Hidden boning',
                },
                pleated: {
                    caption: 'MATERIAL LAB',
                    title: 'Hand Pleated Silk',
                    desc: 'Perforated voile',
                },
                flow: {
                    caption: 'VERNISSAGE NOTES',
                    title: 'Gallery Flow',
                    copy: 'Enter to a central hero gown; material lab sits to the right, curated lookbooks to the left. Soundscape and lighting are layered with a sub-72dB breathing rhythm.',
                    tag: 'LINEAR LUXURY',
                },
                detail: {
                    caption: 'DETAIL',
                    title: 'Paillette Clouds',
                },
            },
        },
    },
    zh: {
        nav: {
            brandLine1: 'NOIR & ÉCLAT',
            brandLine2: 'FLEURLIS',
            links: {
                process: '设计与生产',
                brief: '公司资讯',
                seasonal: '当季系列',
                catalog: '新品目录',
                appointment: '预约到访',
            },
            cta: '预约私享厅',
            language: {
                toggle: '语言',
                zh: '中文',
                en: 'EN',
            },
        },
        footer: {
            tagline: '仅限预约',
            brand: 'FLEURLIS',
            instagram: 'Instagram',
            wechat: '微信',
            email: '邮箱',
            legal: '© 2025 FLEURLIS · 始于 2017 · 苏州',
            icp: '苏ICP备2025229492号-1',
        },
        hero: {
            label: '// INTRODUCTION_THEATRE',
            body: `始于 2008 · 顺河镇工厂
深耕晚礼服定制与批发近二十载
懂品质 · 懂市场 · 懂交付

二十年匠心 · 守护每一道工序
--------------------------------
源头工厂 · 全球供应`,
        },
        ticker: {
            release: '新品发布 · 2025 春夏 · 仅限批发 ·',
            fashionWeek: '上海时装周 · A-12 展位 ·',
            leadTime: '生产周期：14 天 ·',
        },
        seasonal: {
            title: '当季系列',
            season: '2025 春夏',
            series: '系列 {index}',
            lookbook: '查看型录 ->',
        },
        product: {
            title: '上新款式',
            total: '总计：{count}',
            filter: '筛选：全部',
            sort: '排序：日期',
            style: '款号 #{id}',
            login: '登录查看价格',
            add: '加入清单',
        },
        info: {
            designerTitle: '设计师手记',
            designerBody: '我们剔除冗余，没有荷叶边，没有多余装饰，只留下结构与线条。设计就是身体的建筑蓝图。',
            designerSign: '- 创意总监',
            appointmentTitle: '预约到访',
            appointmentBody: '欢迎到展厅亲手触摸面料。我们为批发伙伴提供私享场次。',
            appointmentCta: '预约参观',
            brandTitle: '品牌信息',
            brandStats: {
                est: '创立',
                estValue: '2017',
                hq: '总部',
                hqValue: '苏州',
                clients: '客户',
                clientsValue: '全球 B2B',
                capacity: '产能',
                capacityValue: '50,000/年',
            },
        },
        preview: {
            ref: '档案号：2025-ARCHIVE',
            security: '安全网关',
            heading: '设计 · 生产 · 销售',
            subheading: '始于 2017 · 全流程支持',
            enter: '[ 进入档案 ]',
            ready: '系统就绪',
        },
        atelier: {
            eyebrow: 'THE ATELIER',
            title: '工坊流程 · Linear Precision',
            lede: '以界面化方式呈现高定流程，让客户像看产品规格一样读取面料克重、缝制工序与排期。',
            routineTitle: '标准流程',
            timeline: [
                {
                    title: 'DAY 01 · PATTERN SKETCH',
                    desc: '72 小时线稿与立体裁剪，透明塔夫绸校准版型。',
                },
                {
                    title: 'DAY 04 · MATERIAL CURATION',
                    desc: '香槟白塔夫绸 × 雾面珠光纱，记录温湿度与肌理。',
                },
                {
                    title: 'DAY 09 · FITTING',
                    desc: '三轮试衣 + 激光刺绣定位，保留每次调整的 Data Log。',
                },
            ],
            meta: ['透光测试', '香调匹配', '触感档案'],
            materialLabTitle: 'Material Lab',
            materialLabCopy: '用高调摄影记录用料，并以数字孪生方式同步到 WebGL 丝绸着色器，线上线下色温一致。',
            pills: {
                reflectance: '0.78 · pearl',
                roughness: '0.32 · satin',
                colorDrift: '冷白 ↔ 暖金',
                windProfile: 'Perlin 0.6',
            },
        },
        couture: {
            eyebrow: 'COUTURE & CUSTOM',
            title: '预约 FLEURLIS 试穿',
            copy: '每次量体都在磨砂玻璃的静谧空间进行，Ambient 光模拟拂晓日光。造型顾问按肤色与场合提供 Look Sheet，并同步香氛与面料触感档案。',
            bullets: ['远程体感会议 · 45 分钟', '面料触感档案 + 香氛礼包寄送', '72 小时草图 · 14 天首版试衣'],
            badge: '仅限预约',
            formTitle: '预约专属试穿',
            formCopy: '告诉我们你的城市与理想日期，工坊将为你安排最近的 FLEURLIS Salon。',
            form: {
                city: '城市',
                date: '日期',
                placeholder: 'Shanghai / Paris',
                submit: '提交预约',
            },
        },
        gallery: {
            eyebrow: 'THE GALLERY',
            title: 'Vernissage · 白晕展柜',
            lede: '非对称 Bento Grid 拆分展示礼服、高定配饰与材质特写。1px 线框让卡片悬浮在“工坊迷雾”上，呈现高调摄影高光。',
            cards: {
                halo: {
                    caption: 'LOOK 01 · STRUCTURE',
                    title: 'Architected Halo',
                    desc: '3D 胸衣 · 隐形骨架',
                },
                pleated: {
                    caption: 'MATERIAL LAB',
                    title: 'Hand Pleated Silk',
                    desc: '打孔薄纱',
                },
                flow: {
                    caption: 'VERNISSAGE NOTES',
                    title: 'Gallery Flow',
                    copy: '入口即见中央主礼服，右侧材质试验室，左翼陈列精选 Lookbook。音景与灯光以 72 分贝内的呼吸节奏铺陈。',
                    tag: 'LINEAR LUXURY',
                },
                detail: {
                    caption: 'DETAIL',
                    title: 'Paillette Clouds',
                },
            },
        },
    },
} satisfies Record<SupportedLocale, Record<string, unknown>>
