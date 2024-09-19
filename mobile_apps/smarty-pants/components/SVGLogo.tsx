import React, { useEffect, useRef } from 'react';
import { Animated } from 'react-native';
import Svg, { Path, Rect, Defs, Use } from 'react-native-svg';

const AnimatedRect = Animated.createAnimatedComponent(Rect);

interface SVGLogoProps {
    width?: number;
    height?: number;
    leftBrainColor?: string;
    rightBrainColor?: string;
    centerSquareColor?: string;
    centerSquareBlinkColor?: string;
}

const SVGLogo: React.FC<SVGLogoProps> = ({
                                             width = 40,
                                             height = 40,
                                             leftBrainColor = '#FFF',
                                             rightBrainColor = '#FFF',
                                             centerSquareColor = '#8CA6C9',
                                             centerSquareBlinkColor = '#FFFFFF',
                                         }) => {
    const blinkAnim = useRef(new Animated.Value(0)).current;

    useEffect(() => {
        Animated.loop(
            Animated.sequence([
                Animated.timing(blinkAnim, {
                    toValue: 1,
                    duration: 500,
                    useNativeDriver: false,
                }),
                Animated.timing(blinkAnim, {
                    toValue: 0,
                    duration: 500,
                    useNativeDriver: false,
                }),
            ])
        ).start();
    }, []);

    const centerSquareColorAnimated = blinkAnim.interpolate({
        inputRange: [0, 1],
        outputRange: [centerSquareColor, centerSquareBlinkColor],
    });

    return (
        <Svg
            width={width}
            height={height}
            viewBox="0 0 225 225"
        >
            <Defs>
                <Path id="half" d="M74,1c3.69,0 7.39,0 11.68.37 6.13,2.44 11.68,4.46 17.18,6.6 2.68,1.04 3.26,3.15 3.24,5.97 -.12,22.49 -.06,44.98 -.08,67.47 0,1.26 -.19,2.51 -.29,3.68 -2.02,0 -3.65,0 -5.8,0 0,-3.36 0,-6.59 0,-9.75 -5.12,0 -9.7,0 -14.7,0 0,3.34 0,6.42 0,9.93 -3.49,0 -6.57,0 -9.86,0 0,5.14 0,9.85 0,14.9 3.4,0 6.51,0 9.6,0 0,2.12 0,3.86 0,5.99 -3.35,0 -6.45,0 -9.8,0 0,5.13 0,9.85 0,14.94 3.39,0 6.49,0 9.6,0 0,2.12 0,3.86 0,5.96 -3.32,0 -6.42,0 -9.54,0 0,5.07 0,9.77 0,14.87 3.44,0 6.52,0 10,0 0,3.53 0,6.62 0,9.79 5.01,0 9.6,0 14.71,0 0,-3.42 0,-6.65 0,-10.04 2.26.13 4,0.23 6.1,0.34 0,24.39 .04,48.36 -.11,72.33 0,1.45 -1.25,3.62 -2.51,4.22 -5.69,2.72 -11.6,5 -17.43,7.43 -4.02,0 -8.05,0 -12.67,-0.35 -21.9,-6.56 -31.75,-19 -33.35,-41.24 -.09,-1.28 -2.04,-2.99 -3.48,-3.51 -25.8,-9.37 -37.18,-40.29 -23.36,-64.03 1.39,-2.38 1.57,-3.94 .12,-6.55 -8.58,-15.51 -8.28,-31.09 1,-46.23 6,-9.74 15,-15.77 25.83,-19.04 -.47,-14.4 4.03,-26.65 15.58,-35.11 5.4,-3.96 12.19,-6.02 18.34,-8.95z" />
            </Defs>
            <Use href="#half" fill={leftBrainColor} />
            <Use href="#half" fill={rightBrainColor} transform="scale(-1, 1) translate(-225, 0)" />
            <AnimatedRect
                x="98"
                y="100"
                width="30"
                height="30"
                fill={centerSquareColorAnimated}
            />
        </Svg>
    );
};

export default SVGLogo;