import styled from "styled-components";

export const Bubble = styled.div`
    display: inline-block;
    max-width: 60%;
    padding: 8px 12px;
    border-radius: 15px;
    word-wrap: break-word;
    overflow-wrap: break-word;
    box-sizing: border-box;
    opacity: ${({ pending }) => (pending ? 0.5 : 1)};
`