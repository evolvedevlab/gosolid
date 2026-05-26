import { createSignal, lazy, onMount } from "solid-js";
import { definePage } from "../runtime";

const meta = () => ({
  title: "Home",
});

const Test = lazy(() => import("../Test"));

const Home = (props) => {
  const [count, setCount] = createSignal(0);
  const [isMounted, setMounted] = createSignal(false);

  onMount(() => setMounted(true));

  return (
    <div>
      <p>Count: {count()}</p>
      <button onClick={() => setCount((c) => c + 1)}>Incr</button>
      {isMounted() ? <Test /> : "loading..."}
      {JSON.stringify(props)}
      Bye
    </div>
  );
};

export default definePage(Home, { meta });
