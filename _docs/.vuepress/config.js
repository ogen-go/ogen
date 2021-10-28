
module.exports = {
    title: 'ogen',
    description: 'OpenAPI v3 Code Generation for Go',
    base: '/',

    themeConfig: {
        navbar: [
            {
                text: 'Home',
                link: 'https://ogen.dev'
            },
            {
                text: 'GitHub',
                link: 'https://github.com/ogen-go/ogen'
            }
        ],
        sidebar: [
            {
                text: 'Prologue',
                link: '/prologue'
            },
            {
                text: 'Get Started',
                link: '/start',
                children: [
                    {
                        text: 'Install',
                        link: '/start/#install'
                    },
                ],
            },
            {
                text: 'Features',
                link: '/features',
                children: [
                    {
                        text: 'Generics',
                        link: '/features/#generics'
                    },
                    {
                        text: 'Sum types',
                        link: '/features/#sum-types'
                    },
                ]
            },
        ]
    },

    plugins: [
        ['@vuepress/plugin-search', {
            locales: {
                '/': {
                    placeholder: 'Search'
                }
            }
        }]
    ],
}
