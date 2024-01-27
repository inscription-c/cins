import { defineUserConfig, defaultTheme } from 'vuepress'
import { tocPlugin } from '@vuepress/plugin-toc'
import { searchPlugin } from '@vuepress/plugin-search'
import { backToTopPlugin } from '@vuepress/plugin-back-to-top'

export default defineUserConfig({
    locales: {
        '/': {
            lang: 'en-US',
            title: 'Inscription-Contractualized Protocol',
            description: 'A protocol for the contractualized inscription.',
        },
    },
    theme: defaultTheme({
        navbar: [
            {
                text: 'Docs',
                link: '/',
            },
            {
                text: 'Github',
                link: 'https://github.com/inscription-c/insc',
            }
        ],
        sidebarDepth: 0,
        sidebar: [
            '/README.md',
            {
                text: 'Data Structure',
                children: [
                    '/data-structure/inscription.md',
                    '/data-structure/brc-20-c.md',
                ],
            },
            {
                text: 'Node Guide',
                children: [
                    '/node-guide/installation.md',
                    '/node-guide/run-node.md',
                    '/node-guide/deploy-inscription.md',
                    '/node-guide/http-api-reference.md',
                ],
            },
            '/contributing.md',
        ],
        themePlugins: {
            backToTop: true,
        }
    }),
    plugins: [
        searchPlugin({}),
        backToTopPlugin(),
        tocPlugin({}),
    ],
})
