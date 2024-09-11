// middleware.ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// List of public routes that don't require authentication
const publicRoutes = ['/login', '/auth/google/callback'];

// List of API routes that should use the existing authMiddleware
const apiRoutes = ['/api/'];

export async function middleware(request: NextRequest) {
    const { pathname } = request.nextUrl;

    // Check if the route is public
    if (publicRoutes.some(route => pathname.startsWith(route))) {
        return NextResponse.next();
    }

    // For API routes, we'll continue to use the existing authMiddleware
    if (apiRoutes.some(route => pathname.startsWith(route))) {
        return NextResponse.next();
    }

    // For other routes, check for the auth token in cookies
    const token = request.cookies.get('auth_token')?.value;

    if (!token) {
        // Redirect to login if there's no token
        return NextResponse.redirect(new URL('/login', request.url));
    }

    // If there's a token, allow the request to proceed
    return NextResponse.next();
}

export const config = {
    matcher: [
        /*
         * Match all request paths except for the ones starting with:
         * - _next/static (static files)
         * - _next/image (image optimization files)
         * - favicon.ico (favicon file)
         */
        '/((?!_next/static|_next/image|favicon.ico).*)',
    ],
};