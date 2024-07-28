// File: /utils/datasources.ts

import {AvailableDatasource} from '@/types/datasource';

export const availableDatasources: AvailableDatasource[] = [
    {
        id: 'slack',
        name: 'Slack',
        imageUrl: '/images/slack_icon.webp',
        description: 'Connect and sync your Slack workspace',
        configurationUrl: '/datasources/slack/add',
    },
    {
        id: 'confluence',
        name: 'Confluence',
        imageUrl: '/images/confluence_icon.png',
        description: 'Integrate your Confluence knowledge base',
        configurationUrl: '#',
    },
    {
        id: 'gmail',
        name: 'Gmail',
        imageUrl: '/images/gmail_icon.png',
        description: 'Index your email content for asking LLM questions',
        configurationUrl: '#',
    },
    {
        id: 'postgres',
        name: 'PostgreSQL',
        imageUrl: '/images/postgres_icon.png',
        description: 'Index directly from your PostgreSQL database',
        configurationUrl: '#',
    },
    {
        id: 'jira',
        name: 'Jira',
        imageUrl: '/images/jira_icon.png',
        description: 'Add your Jira issues and projects',
        configurationUrl: '#',
    },
    {
        id: 'gdrive',
        name: 'Google Drive',
        imageUrl: '/images/gdrive_icon.png',
        description: 'Index your Google Drive documents',
        configurationUrl: '#',
    },
    {
        id: 'web_crawler',
        name: 'Web Crawler',
        imageUrl: '/images/web_crawler_icon.png',
        description: 'Crawl and index web pages',
        configurationUrl: '#',
    },
    {
        id: 'microsoft_office_team',
        name: 'Microsoft Teams',
        imageUrl: '/images/microsoft_team_icon.webp',
        description: 'Index your Microsoft Teams messages',
        configurationUrl: '#',
    },
    {
        id: 'file_upload',
        name: 'Upload Local Files',
        imageUrl: '/images/file_upload_icon.png',
        description: 'Upload files to be indexed',
        configurationUrl: '#',
    },
    {
        id: 'rest_api',
        name: 'REST API',
        imageUrl: '/images/api_icon.png',
        description: 'Upload data via REST API',
        configurationUrl: '#',
    },
];

export function getDatasourceById(id: string): AvailableDatasource | undefined {
    return availableDatasources.find(datasource => datasource.id === id);
}