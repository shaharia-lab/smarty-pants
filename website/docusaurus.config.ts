import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';
import * as path from "node:path";

const config: Config = {
  title: 'SmartyPants',
  tagline: 'Democratizing Generative AI. Start using Generative AI without any pre-requisite skills and domain knowledge.',
  favicon: 'img/favicon.svg',

  // Set the production url of your site here
  url: 'https://smarty-pants.shaharialab.com',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'shaharia-lab', // Usually your GitHub org/user name.
  projectName: 'smarty-pants', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  plugins: [
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'about',
        path: 'about',
        routeBasePath: 'about',
        sidebarPath: require.resolve('./aboutSidebar.js'),
      },
    ],
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/shaharia-lab/smarty-pants/tree/main/website/docs/',
        },
        blog: {
          showReadingTime: true,
          feedOptions: {
            type: ['rss', 'atom'],
            xslt: true,
          },
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/shaharia-lab/smarty-pants/tree/main/website/blog/',
          // Useful options to enforce blogging best practices
          onInlineTags: 'warn',
          onInlineAuthors: 'warn',
          onUntruncatedBlogPosts: 'warn',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: 'img/smartypants-social-card.png',
    navbar: {
      title: 'SmartyPants',
      logo: {
        alt: 'SmartyPants Logo',
        src: 'img/logo_light.svg',
      },
      items: [
        {
          href: '/about',
          label: 'About SmartyPants',
          position: 'left',
        },
        {
          href: '/docs/category/documentations',
          label: 'Documentations',
          position: 'left',
        },
        {to: '/blog', label: 'Blog', position: 'left'},
        {
          href: 'https://github.com/shaharia-lab/smarty-pants',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'About SmartyPants',
          items: [
            {
              label: 'About SmartyPants',
              to: '/docs/about/intro',
            },
          ],
        },
        {
          title: 'Docs',
          items: [
            {
              label: 'Documentation',
              to: '/docs/documentations/installation',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Stack Overflow',
              href: 'https://stackoverflow.com/questions/tagged/smarty-pants',
            },
            {
              label: 'Discord',
              href: 'https://discordapp.com/invite/smarty-pants',
            },
            {
              label: 'Twitter',
              href: 'https://twitter.com/shaharia-lab',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'Blog',
              to: '/blog',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/shaharia-lab/smarty-pants',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} <a href="https://shaharialab.com/open-source?utm_source=smarty_pants_website">Shaharia Lab OÜ</a>. An Open Source initiative. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
