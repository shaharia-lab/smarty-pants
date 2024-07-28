import './globals.css'
import {Inter} from 'next/font/google'
import React from "react";

const inter = Inter({subsets: ['latin']})

export const metadata = {
    title: 'SmartyPants AI',
    description: 'Discover search differently',
}

type RootLayoutProps = Readonly<{
    children: React.ReactNode
}>

export default function RootLayout({children}: RootLayoutProps) {
    return (
        <html lang="en" className={`${inter.className} h-full`}>
        <body className="h-full">
        {children}
        </body>
        </html>
    )
}