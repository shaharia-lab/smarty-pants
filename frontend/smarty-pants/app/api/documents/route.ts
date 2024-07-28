import {NextRequest, NextResponse} from 'next/server';

export const runtime = 'edge';

export async function GET(request: NextRequest) {
    const searchParams = request.nextUrl.searchParams;
    const page = searchParams.get('page') || '1';
    const status = searchParams.get('status') || '';
    const limit = searchParams.get('limit') || '10';

    // Here you would typically fetch data from your database or external API
    // For this example, we'll use mock data
    const mockDocuments = [
        {uuid: "1", title: "Document 1", status: "pending", created_at: "2024-07-06T22:07:01.386319Z"},
        {uuid: "2", title: "Document 2", status: "processed", created_at: "2024-07-07T10:15:30.123456Z"},
        // ... add more mock documents as needed
    ];

    const filteredDocuments = status
        ? mockDocuments.filter(doc => doc.status === status)
        : mockDocuments;

    const startIndex = (Number(page) - 1) * Number(limit);
    const endIndex = startIndex + Number(limit);
    const paginatedDocuments = filteredDocuments.slice(startIndex, endIndex);

    return NextResponse.json({
        documents: paginatedDocuments,
        total: filteredDocuments.length,
    });
}